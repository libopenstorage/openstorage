// gcc fuse.c -D_FILE_OFFSET_BITS=64 -lfuse -lulockmgr -o fuse

#define _GNU_SOURCE
#define _FILE_OFFSET_BITS 64
#define FUSE_USE_VERSION 26

#ifdef HAVE_CONFIG_H
#include <config.h>
#endif

#define EXPERIMENTAL_

#ifdef EXPERIMENTAL_

#define _GNU_SOURCE

#include <fuse.h>
#include <ulockmgr.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdbool.h>
#include <fcntl.h>
#include <dirent.h>
#include <errno.h>
#include <sys/time.h>
#ifdef HAVE_SETXATTR
#include <sys/xattr.h>
#endif

#define MAX_DESC 4096
#define MAX_LAYERS 16

static char *union_src;

struct union_fs {
	char *layers[MAX_LAYERS];
};

struct union_fs *ufs = NULL;

struct descriptor {
	char name[PATH_MAX];
	int fd;
} descriptors[MAX_DESC];

struct graph_dirp {
	DIR *dp;
	struct dirent *entry;
	off_t offset;
};

static void descriptors_init() {
	int i;
	for (i=0; i<MAX_DESC; ++i) {
		descriptors[i].fd = -1;
		descriptors[i].name[0] = 0;
	}
}

static int find_descriptor(const char* path) {
	int i;
	for (i=0; i<MAX_DESC; ++i) {
		if(!strcmp(descriptors[i].name, path)) {
			return descriptors[i].fd;
		}
	}
	return -1;
}

static int register_fd(const char* path, int fd) {
	int i;
	for (i=0; i<MAX_DESC; ++i) {
		if(descriptors[i].fd == -1) {
			descriptors[i].fd = fd;
			snprintf(descriptors[i].name, PATH_MAX, "%s", path);
			return fd;
		}
	}
	return -1;
}

static char *real_path(const char *path, bool create_mode)
{
	char *r = NULL;

	errno = 0;

	if (ufs != NULL) {
		int i;
		int ret;
		struct stat st;

		for (i = 0; ufs->layers[i]; i++) {
			asprintf(&r, "%s%s", ufs->layers[i], path);
			if (!r) {
				errno = ENOMEM;
				fprintf(stderr, "Warning, cannot allocate memory\n");
				break;
			}

			ret = lstat(r, &st);
			if (ret == 0) {
				break;
			}

			free(r);
			r = NULL;
		}

		if (create_mode && ufs->layers[0]) {
			asprintf(&r, "%s%s", ufs->layers[i], path);
			if (!r) {
				errno = ENOMEM;
				fprintf(stderr, "Warning, cannot allocate memory\n");
			}
		}
	}

	if (!errno && !r) {
		fprintf(stderr, "Warning, requested path not found: %s\n", path);
		errno = ENOENT;
	}

	return r;
}

static void free_path(char *path)
{
	free(path);
}

static int maybe_open(const char* path, int flags, int mode) {
	int fd;
	int ret;
	char *rp = NULL;

	rp = real_path(path, (flags & O_CREAT ? true : false));
	if (!rp) {
		goto done;
	}

	fd = find_descriptor(rp);
	if (fd != -1) {
		goto done;
	}

	int fixed_flags = (flags & (~O_WRONLY) & (~O_RDONLY)) | O_RDWR;

	fd = open(rp, fixed_flags, mode);
	if (fd==-1) {
		fd = open(rp, flags, mode);
	}

	if (fd==-1) {
		goto done;
	}

	ret = register_fd(rp, fd);
	if (ret==-1)  {
		fprintf(stderr, "Warning, error while registering FD for %s.\n", rp);
		close(fd);
		fd = -1;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return fd;
}

static int graph_opendir(const char *path, struct fuse_file_info *fi)
{
	int res = 0;
	char *rp = NULL;
	struct graph_dirp *d = malloc(sizeof(struct graph_dirp));

	if (d == NULL) {
		res = -ENOMEM;
		goto done;
	}

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	d->dp = opendir(rp);
	if (d->dp == NULL) {
		res = -errno;
		free(d);
		goto done;
	}
	d->offset = 0;
	d->entry = NULL;

	fi->fh = (unsigned long) d;

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static inline struct graph_dirp *get_dirp(struct fuse_file_info *fi)
{
	return (struct graph_dirp *) (uintptr_t) fi->fh;
}

static int graph_readdir(const char *path, void *buf, fuse_fill_dir_t filler,
		off_t offset, struct fuse_file_info *fi)
{
	struct graph_dirp *d = get_dirp(fi);
	(void) path;

	if (offset != d->offset) {
		seekdir(d->dp, offset);
		d->entry = NULL;
		d->offset = offset;
	}

	while (1) {
		struct stat st;
		off_t nextoff;

		if (!d->entry) {
			d->entry = readdir(d->dp);
			if (!d->entry) {
				break;
			}
		}

		memset(&st, 0, sizeof(st));
		st.st_ino = d->entry->d_ino;
		st.st_mode = d->entry->d_type << 12;
		nextoff = telldir(d->dp);
		if (filler(buf, d->entry->d_name, &st, nextoff)) {
			break;
		}

		d->entry = NULL;
		d->offset = nextoff;
	}

	return 0;
}

static int graph_releasedir(const char *path, struct fuse_file_info *fi)
{
	struct graph_dirp *d = get_dirp(fi);
	(void) path;

	closedir(d->dp);
	free(d);

	return 0;
}

static int graph_getattr(const char *path, struct stat *stbuf)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = lstat(rp, stbuf);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_access(const char *path, int mask)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = access(rp, mask);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}
	return res;
}

static int graph_readlink(const char *path, char *buf, size_t size)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = readlink(rp, buf, size - 1);
	if (res == -1) {
		res = -errno;
		goto done;
	}
	buf[res] = '\0';
	res = 0;

done:
	if (rp) {
		free_path(rp);
	}
	return res;
}

static int graph_unlink(const char *path)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = unlink(rp);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_rmdir(const char *path)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = rmdir(rp);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_symlink(const char *from, const char *to)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(to, true);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = symlink(from, rp);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_rename(const char *from, const char *to)
{
	int res = 0;
	char *from_rp = NULL;
	char *to_rp = NULL;

	from_rp = real_path(from, false);
	if (!from_rp) {
		res = -errno;
		goto done;
	}

	to_rp = real_path(to, true);
	if (!to_rp) {
		res = -errno;
		goto done;
	}

	res = rename(from_rp, to_rp);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (from_rp) {
		free_path(from_rp);
	}

	if (to_rp) {
		free_path(to_rp);
	}

	return res;
}

static int graph_link(const char *from, const char *to)
{
	int res = 0;
	char *from_rp = NULL;
	char *to_rp = NULL;

	from_rp = real_path(from, false);
	if (!from_rp) {
		res = -errno;
		goto done;
	}

	to_rp = real_path(to, true);
	if (!to_rp) {
		res = -errno;
		goto done;
	}

	res = link(from_rp, to_rp);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (from_rp) {
		free_path(from_rp);
	}

	if (to_rp) {
		free_path(to_rp);
	}

	return res;
}

static int graph_chmod(const char *path, mode_t mode)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = chmod(rp, mode);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_chown(const char *path, uid_t uid, gid_t gid)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = lchown(rp, uid, gid);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_truncate(const char *path, off_t size)
{
	int res = 0;

	int fd = maybe_open(path, O_RDWR, 0777);
	if (fd == -1) {
		errno = ENOENT;
		res = -ENOENT;
		goto done;
	}

	res = ftruncate(fd, size);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:

	return res;
}

static int graph_utimens(const char *path, const struct timespec ts[2])
{
	struct timeval tv[2];
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	tv[0].tv_sec = ts[0].tv_sec;
	tv[0].tv_usec = ts[0].tv_nsec / 1000;
	tv[1].tv_sec = ts[1].tv_sec;
	tv[1].tv_usec = ts[1].tv_nsec / 1000;

	res = utimes(rp, tv);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_open(const char *path, struct fuse_file_info *fi)
{
	int res = 0;
	int fd;

	fd = maybe_open(path, fi->flags, 0777);
	if (fd == -1) {
		res = -errno;
		goto done;
	}

	fi->fh = fd;

done:

	return res;
}

static int graph_create(const char *path, mode_t mode, struct fuse_file_info *fi)
{
	int res = 0;
	int fd;

	fd = maybe_open(path, fi->flags, mode);
	if (fd == -1) {
		res = -errno;
		goto done;
	}

	fi->fh = fd;

done:

	return res;
}

static int graph_mknod(const char *path, mode_t mode, dev_t rdev)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, true);
	if (!rp) {
		res = -errno;
		goto done;
	}

	if (S_ISFIFO(mode)) {
		res = mkfifo(rp, mode);
	} else {
		res = mknod(rp, mode, rdev);
	}

	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_mkdir(const char *path, mode_t mode)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, true);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = mkdir(rp, mode);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_fgetattr(const char *path, struct stat *stbuf,
		struct fuse_file_info *fi)
{
	int res = 0;
	(void) path;

	res = fstat(fi->fh, stbuf);
	if (res == -1) {
		return -errno;
	}

	return 0;
}

static int graph_ftruncate(const char *path, off_t size,
		struct fuse_file_info *fi)
{
	int res;
	(void) path;

	res = ftruncate(fi->fh, size);
	if (res == -1) {
		return -errno;
	}

	return 0;
}

static int graph_read(const char *path, char *buf, size_t size, off_t offset,
		struct fuse_file_info *fi)
{
	int res;
	(void) path;

	res = pread(fi->fh, buf, size, offset);
	if (res == -1) {
		res = -errno;
	}

	return res;
}

static int graph_write(const char *path, const char *buf, size_t size,
		off_t offset, struct fuse_file_info *fi)
{
	int res;
	(void) path;

	res = pwrite(fi->fh, buf, size, offset);
	if (res == -1) {
		res = -errno;
	}

	return res;
}

static int graph_statfs(const char *path, struct statvfs *stbuf)
{
	int res = 0;

	res = statvfs(union_src, stbuf);
	if (res == -1) {
		res = -errno;
	}

	return res;
}

static int graph_flush(const char *path, struct fuse_file_info *fi)
{
	(void) path;

	return 0;
}

static int graph_release(const char *path, struct fuse_file_info *fi)
{
	(void) path;
	// intentionally omit closing fi->fh.

	return 0;
}

static int graph_fsync(const char *path, int isdatasync,
		struct fuse_file_info *fi)
{
	int res;
	(void) path;

#ifndef HAVE_FDATASYNC
	(void) isdatasync;
#else
	if (isdatasync)
		res = fdatasync(fi->fh);
	else
#endif
		res = fsync(fi->fh);
	if (res == -1)
		return -errno;

	return 0;
}

#ifdef HAVE_SETXATTR
/* xattr operations are optional and can safely be left unimplemented */
static int graph_setxattr(const char *path, const char *name, const char *value,
		size_t size, int flags)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = lsetxattr(rp, name, value, size, flags);
	if (res == -1) {
		res = -errno;
		goto done;
	}
	res = 0;

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_getxattr(const char *path, const char *name, char *value,
		size_t size)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = lgetxattr(rp, name, value, size);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_listxattr(const char *path, char *list, size_t size)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = llistxattr(rp, list, size);
	if (res == -1) {
		res = -errno;
		goto done;
	}

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

static int graph_removexattr(const char *path, const char *name)
{
	int res = 0;
	char *rp = NULL;

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	int res = lremovexattr(rp, name);
	if (res == -1) {
		ret = -errno;
		goto done;
	}
	res = 0;

done:
	if (rp) {
		free_path(rp);
	}

	return res;
}

#endif /* HAVE_SETXATTR */

static int graph_lock(const char *path, struct fuse_file_info *fi, int cmd,
		struct flock *lock)
{
	(void) path;

	return ulockmgr_op(fi->fh, cmd, lock, &fi->lock_owner,
			sizeof(fi->lock_owner));
}

static struct fuse_operations graph_oper = {
	.getattr	= graph_getattr,
	.fgetattr	= graph_fgetattr,
	.access		= graph_access,
	.readlink	= graph_readlink,
	.opendir	= graph_opendir,
	.readdir	= graph_readdir,
	.releasedir	= graph_releasedir,
	.mknod		= graph_mknod,
	.mkdir		= graph_mkdir,
	.symlink	= graph_symlink,
	.unlink		= graph_unlink,
	.rmdir		= graph_rmdir,
	.rename		= graph_rename,
	.link		= graph_link,
	.chmod		= graph_chmod,
	.chown		= graph_chown,
	.truncate	= graph_truncate,
	.ftruncate	= graph_ftruncate,
	.utimens	= graph_utimens,
	.create		= graph_create,
	.open		= graph_open,
	.read		= graph_read,
	.write		= graph_write,
	.statfs		= graph_statfs,
	.flush		= graph_flush,
	.release	= graph_release,
	.fsync		= graph_fsync,
#ifdef HAVE_SETXATTR
	.setxattr	= graph_setxattr,
	.getxattr	= graph_getxattr,
	.listxattr	= graph_listxattr,
	.removexattr= graph_removexattr,
#endif
	.lock		= graph_lock,

	.flag_nullpath_ok = 1,
};

int start_unionfs(char *src_path, char *mount_path)
{
	char *argv[4];

	union_src = strdup(src_path);
	if (!union_src) {
		return -errno;
	}

	descriptors_init();
	umask(0);

	argv[0] = "graph";
	argv[1] = mount_path;
	argv[2] = "-f";

	return fuse_main(3, argv, &graph_oper, NULL);
}

int alloc_unionfs(char *layer_path, char *id)
{
	char link[PATH_MAX];
	char *layer;
	int res = 0;
	int i;

	ufs = (struct union_fs *)calloc(1, sizeof(struct union_fs));
	if (!ufs) {
		res = -errno;
		goto done;
	}

	printf("Unifying %s\n", layer_path);

	for (i = 0, layer = layer_path; layer && (i < MAX_LAYERS); i++) {
		char *parent = NULL;

		ufs->layers[i] = strdup(layer);
		if (!ufs->layers[i]) {
			res = -errno;
			goto done;
		}

		printf("Added Layer: %d: %s\n", i, layer);

		asprintf(&parent, "%s/_parent", layer);
		if (!parent) {
			res = -errno;
			goto done;
		}

		res = readlink(parent, link, sizeof(link));
		if (res != -1) {
			layer = link;
		} else {
			res = 0;
			layer = NULL;
		}

		free(parent);
	}

done:
	if (res != 0) {
		if (ufs) {
			free(ufs);
		}
	} else {
		errno = 0;
	}

	return res;
}

int release_unionfs(char *id)
{
	return 0;
}


/*
int main()
{
	start_unionfs("/var/lib/openstorage/fuse/physical", "/var/lib/openstorage/fuse/virtual");
}
*/

#endif
