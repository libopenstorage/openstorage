// gcc fuse.c -DiFILE_OFFSET_BITS=64 -lfuse -lulockmgr -lithread -o fuse

#define _GNU_SOURCE
#define _FILE_OFFSET_BITS 64
#define FUSE_USE_VERSION 26

#ifdef HAVE_CONFIG_H
#include <config.h>
#endif

// XXX FIXME
// #define EXPERIMENTAL_ 1

#ifdef EXPERIMENTAL_

#define _GNU_SOURCE

#include <fuse.h>
#include <pthread.h>
#include <ulockmgr.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <libgen.h>
#include <unistd.h>
#include <stdbool.h>
#include <fcntl.h>
#include <dirent.h>
#include <errno.h>
#include <sys/time.h>
#ifdef HAVE_SETXATTR
#include <sys/xattr.h>
#endif

#include "hash.h"

#define MAX_DESC 4096
#define MAX_LAYERS 64
#define MAX_INSTANCES 128

struct union_fs {
	char *layers[MAX_LAYERS];
	pthread_mutex_t lock;
	char id[256];
	bool available;
};

static char *union_src;
static pthread_mutex_t ufs_lock;
static struct union_fs ufs_instances[MAX_INSTANCES];
static hashtable_t *ufs_hash;

struct descriptor 
{
	char name[PATH_MAX];
	int fd;
} descriptors[MAX_DESC];

struct graph_dirp 
{
	DIR *dp;
	struct dirent *entry;
	off_t offset;
};

static void trace(const char *fn, const char *path)
{
	// fprintf(stderr, "unionfs operation: %s on %s\n", fn, path);
}

static void lock_ufs(struct union_fs *ufs)
{
	pthread_mutex_lock(&ufs->lock);
}

static void unlock_ufs(struct union_fs *ufs)
{
	pthread_mutex_unlock(&ufs->lock);
}

static struct union_fs *get_ufs(const char *path, char **new_path)
{
	struct union_fs *ufs = NULL;
	char *p, *tmp_path = NULL;
	int i, id;

	*new_path = NULL;

	tmp_path = strdup(path + 1);
	if (!tmp_path) {
		fprintf(stderr, "Warning, cannot allocate memory.\n");
		goto done;
	}
	p = strchr(tmp_path, '/');
	if (p) *p = 0;

	ufs = ht_get(ufs_hash, tmp_path);
	if (!ufs) {
		goto done;
	}

	if (!ufs->available) {
		*new_path = strchr(path+1, '/');
		if (!*new_path) {
			// Must be a request for root.
			*new_path = "/";
		}
	} else {
		ufs = NULL; 
	}

done:
	if (tmp_path) {
		free(tmp_path);
	}

	return ufs;
}

static void descriptors_init() 
{
	int i;
	for (i=0; i<MAX_DESC; ++i) {
		descriptors[i].fd = -1;
		descriptors[i].name[0] = 0;
	}
}

static int find_descriptor(const char* path) 
{
	int i;
	for (i=0; i<MAX_DESC; ++i) {
		if(!strcmp(descriptors[i].name, path)) {
			return descriptors[i].fd;
		}
	}
	return -1;
}

static int register_fd(const char* path, int fd) 
{
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
	char file[PATH_MAX];
	char *dir;
	struct union_fs *ufs = NULL;
	char *fixed_path = NULL;

	if (!strcmp(path, "/")) {
		// This is a request for the root virtual path.  There are only
		// union FS volumes at this location and no specific union FS context.
		r = strdup(union_src);
		goto done;
	}

	ufs = get_ufs(path, &fixed_path);
	if (!ufs) {
		// Assume the request is for a raw physical layer.
		asprintf(&r, "%s%s", union_src, path);
		goto done;
	}

	lock_ufs(ufs);

	strncpy(file, fixed_path, sizeof(file));
	dir = dirname(file);

	errno = 0;

	if (ufs != NULL) {
		int base_layer = -1;
		int i;
		int ret;
		struct stat st;

		for (i = 0; ufs->layers[i]; i++) {
			asprintf(&r, "%s%s", ufs->layers[i], fixed_path);
			if (!r) {
				errno = ENOMEM;
				fprintf(stderr, "Warning, cannot allocate memory\n");
				goto done;
			}

			ret = lstat(r, &st);
			if (ret == 0) {
				// Found the file.
				goto done;
			}

			// See if this layer contains the parent directory.  We give
			// preference to the upper layers.
			if (base_layer == -1) {
				char *tmp_r;
				asprintf(&tmp_r, "%s%s", ufs->layers[i], dir);
				if (!r) {
					errno = ENOMEM;
					fprintf(stderr, "Warning, cannot allocate memory\n");
					goto done;
				}

				ret = lstat(tmp_r, &st);
				if (ret == 0) {
					// This layer can be used to create the file.
					base_layer = i;
				}
				free(tmp_r);
			}

			free(r);
			r = NULL;
		}

		// If we did not find the file and create mode was requested, construct
		// a file path in the appropriate layer.	
		if (!r && create_mode && ufs->layers[0]) {
			if (base_layer == -1) {
				fprintf(stderr, "Warning, create mode requested on %s, "
						"but no layer could be found that could create this file\n", fixed_path);
				errno = ENOENT;
			} else {
				asprintf(&r, "%s%s", ufs->layers[base_layer], fixed_path);
				if (!r) {
					fprintf(stderr, "Warning, cannot allocate memory\n");
					errno = ENOMEM;
				}
			}
		}
	} else {
		fprintf(stderr, "Warning, union FS not yet initialized.  Cannot access: %s\n", fixed_path);
		errno = ENOENT;
	}

done:

	if (ufs) {
		unlock_ufs(ufs);
	}

	return r;
}

static void free_path(char *path)
{
	free(path);
}

// Find a file in the FD cache.
static int maybe_open(const char* path, int flags, int mode) 
{
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
		if (flags & O_CREAT) {
			fprintf(stderr, "Warning, failed to create %s (errno=%d)\n", rp, errno);
		}
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

	// trace(__func__, path);

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

// This does the bulk of unifying entries from the various layers.
// It has to make sure dup entries are avoided.
static int graph_readdir(const char *path, void *buf, fuse_fill_dir_t filler,
		off_t offset, struct fuse_file_info *fi)
{
	int res = 0;
	struct union_fs *ufs = NULL;
	char *fixed_path = NULL;
	struct stat st;
	int i;

	// trace(__func__, path);

	if (!strcmp(path, "/")) {
		// List valid union FS paths.

		// XXX do we need to lock here?
		pthread_mutex_lock(&ufs_lock);
		{
			for (i = 0; i < MAX_INSTANCES; i++) {
				if (!ufs_instances[i].available) {
					char phys_path[PATH_MAX];
					char d_name[8];

					snprintf(phys_path, sizeof(phys_path), "%s/%s",
						union_src, ufs_instances[i].id);
					stat(phys_path, &st);

					sprintf(d_name, "%s", ufs_instances[i].id);
					if (filler(buf, d_name, &st, 0)) {
						fprintf(stderr, "Warning, Filler too full on root.\n");
						break;
					}
				}
			}
		}
		pthread_mutex_unlock(&ufs_lock);

		goto done;
	}

	ufs = get_ufs(path, &fixed_path);
	if (!ufs) {
		errno = ENOENT;
		res = -errno;
		// fprintf(stderr, "Warning, no valid union FS for %s\n", path);
		goto done;
	}

	lock_ufs(ufs);

	for (i = 0; ufs->layers[i]; i++) {
		char *rp = NULL;
		int ret;

		asprintf(&rp, "%s%s", ufs->layers[i], fixed_path);
		if (!rp) {
			errno = ENOMEM;
			fprintf(stderr, "Warning, cannot allocate memory\n");
			break;
		}

		ret = lstat(rp, &st);
		if (ret == 0) {
			DIR *dp;

			dp = opendir(rp);
			if (!dp) {
				fprintf(stderr, "Warning, %s not a directory.\n", rp);
				free(rp);
				continue;
			}

			while (true) {
				struct dirent *entry = NULL;
				entry = readdir(dp);
				if (!entry) {
					break;
				}

				if (strcmp(".", entry->d_name) == 0 ||
					strcmp("..", entry->d_name) == 0 || 
					strcmp(".unionfs.parent", entry->d_name) == 0) {
					continue;
				}

				memset(&st, 0, sizeof(st));
				// st.st_ino = entry->d_ino;
				// st.st_mode = entry->d_type << 12;

				// XXX FIXME - make use of tge bext off feature in fuse.
				if (filler(buf, entry->d_name, &st, 0)) {
					fprintf(stderr, "Warning, Filler too full on %s.\n", rp);
					break;
				}
			}
		}

		free(rp);
	}

done:

	if (ufs) {
		unlock_ufs(ufs);
	}

	return res;
}

static int graph_releasedir(const char *path, struct fuse_file_info *fi)
{
	struct graph_dirp *d = get_dirp(fi);
	(void) path;

	// trace(__func__, path);

	closedir(d->dp);
	free(d);

	return 0;
}

static int graph_getattr(const char *path, struct stat *stbuf)
{
	int res = 0;
	char *rp = NULL;

	// trace(__func__, path);

	rp = real_path(path, false);
	if (!rp) {
		res = -errno;
		goto done;
	}

	res = lstat(rp, stbuf);
	stbuf->st_ino = 0;
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

	// trace(__func__, path);

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

	// trace(__func__, path);

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

	// trace(__func__, path);

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

	// trace(__func__, path);

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

	// trace(__func__, from);
	// trace(__func__, to);

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

	// trace(__func__, from);
	// trace(__func__, to);

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

	// trace(__func__, from);
	// trace(__func__, to);

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

	// trace(__func__, path);

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

	// trace(__func__, path);

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

	// trace(__func__, path);

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

	trace(__func__, path);

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

	trace(__func__, path);

	fd = maybe_open(path, fi->flags, mode);
	if (fd == -1) {
		res = -errno;
		goto done;
	}

	fi->fh = fd;

done:

	return res;
}

static int graph_mkdir(const char *path, mode_t mode)
{
	int res = 0;
	char *rp = NULL;

	// trace(__func__, path);

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

static int graph_mknod(const char *path, mode_t mode, dev_t rdev)
{
	int res = 0;
	char *rp = NULL;

	// trace(__func__, path);

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

static int graph_fgetattr(const char *path, struct stat *stbuf,
		struct fuse_file_info *fi)
{
	int res = 0;
	(void) path;

	// trace(__func__, path);

	res = fstat(fi->fh, stbuf);
	stbuf->st_ino = 0;
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

	// trace(__func__, path);

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

	// trace(__func__, path);

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
	int i;

	union_src = strdup(src_path);
	if (!union_src) {
		return -errno;
	}

	pthread_mutex_init(&ufs_lock, NULL);

	for (i = 0; i < MAX_INSTANCES; i++) {
		pthread_mutex_init(&ufs_instances[i].lock, NULL);

		ufs_instances[i].available = true;
	}

	ufs_hash = ht_create( 65536 );

	descriptors_init();
	umask(0);

	argv[0] = "graph";
	argv[1] = mount_path;
	argv[2] = "-f";

	return fuse_main(3, argv, &graph_oper, NULL);
}

int alloc_unionfs(char *layer_path, char *id)
{
	struct union_fs *ufs = NULL;
	char link[PATH_MAX];
	char *layer;
	int res = 0;
	int i;

	pthread_mutex_lock(&ufs_lock);
	{
		for (i = 0; i < MAX_INSTANCES; i++) {
			if (ufs_instances[i].available) {
				ufs_instances[i].available = false;
				ufs = &ufs_instances[i];
				break;
			}
		}
	}
	pthread_mutex_unlock(&ufs_lock);

	if (!ufs) {
		errno = ENOMEM;
		res = -errno;
		printf("Warning, no more union FS instances available.\n");
		goto done;
	}

	lock_ufs(ufs);

	memset(ufs, 0, sizeof(struct union_fs));
	strncpy(ufs->id, id, sizeof(ufs->id));
	ht_set(ufs_hash, id, ufs);

	for (i = 0, layer = layer_path; layer && (i < MAX_LAYERS); i++) {
		char *parent = NULL;

		ufs->layers[i] = strdup(layer);
		if (!ufs->layers[i]) {
			res = -errno;
			goto done;
		}

		asprintf(&parent, "%s/.unionfs.parent", layer);
		if (!parent) {
			res = -errno;
			goto done;
		}

		memset(link, 0, sizeof(link));
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

	if (ufs) {
		unlock_ufs(ufs);
	}

	return res;
}

int release_unionfs(char *id)
{
	int i;

	pthread_mutex_lock(&ufs_lock);
	{
		for (i = 0; i < MAX_INSTANCES; i++) {
			if (!ufs_instances[i].available && !strcmp(ufs_instances[i].id, id)) {
				ufs_instances[i].available = true;
				break;
			}
		}
	}
	pthread_mutex_unlock(&ufs_lock);

	return 0;
}

/*
int main()
{
   start_unionfs("/var/lib/openstorage/fuse/physical", "/var/lib/openstorage/fuse/virtual");
}
*/

#endif
