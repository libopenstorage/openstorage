// gcc unionfs.c layer.c hash.c -DEXPERIMENTAL_ -DSTANDALONE_ -DiFILE_OFFSET_BITS=64 -lfuse -lulockmgr -lpthread -o unionfs

#ifdef EXPERIMENTAL_

#define _GNU_SOURCE
#define _FILE_OFFSET_BITS 64
#define FUSE_USE_VERSION 26

#ifdef HAVE_CONFIG_H
#include <config.h>
#endif

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
#include <sys/types.h>
#ifdef HAVE_SETXATTR
#include <sys/xattr.h>
#endif

#include "hash.h"
#include "layer.h"

struct graph_dirp 
{
	struct dirent *entry;
	off_t offset;
};

static int graph_opendir(const char *path, struct fuse_file_info *fi)
{
	int res = 0;
#if 0
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
#endif
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
#if 0
	struct union_fs *ufs = NULL;
	char *fixed_path = NULL;
	struct stat st;
	int i;

	if (!strcmp(path, "/")) {
		// List all layers.
		for (i = 0; i < MAX_INSTANCES; i++) {
			if (!ufs_heap[i].available) {
				char phys_path[PATH_MAX];
				char d_name[8];

				snprintf(phys_path, sizeof(phys_path), "%s/%s",
					union_src, ufs_heap[i].id);
				stat(phys_path, &st);

				sprintf(d_name, "%s", ufs_heap[i].id);
				if (filler(buf, d_name, &st, 0)) {
					fprintf(stderr, "Warning, Filler too full on root.\n");
					break;
				}
			}
		}

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

#endif
	return res;
}

static int graph_releasedir(const char *path, struct fuse_file_info *fi)
{
#if 0
	struct graph_dirp *d = get_dirp(fi);
	(void) path;

	closedir(d->dp);
	free(d);
#endif
	return 0;
}

static int graph_getattr(const char *path, struct stat *stbuf)
{
	int res = 0;
	struct inode *inode = NULL;

	inode = ref_inode(path, false, 0);	
	if (!inode) {
		res = -errno;
		goto done;
	}

	memset(stbuf, 0, sizeof(struct stat));

	stbuf->st_mode = inode->mode;
	stbuf->st_nlink = inode->nlink;
	stbuf->st_uid = inode->uid;
	stbuf->st_gid = inode->gid;
	stbuf->st_size = inode->size;
	stbuf->st_atime = inode->atime;
	stbuf->st_mtime = inode->mtime;
	stbuf->st_ctime = inode->ctime;
	stbuf->st_ino = 0;

done:
	if (inode) {
		deref_inode(inode);
	}

	return res;
}

static int graph_access(const char *path, int mask)
{
	int res = 0;
	struct inode *inode = NULL;

	inode = ref_inode(path, false, 0);	
	if (!inode) {
		res = -errno;
		goto done;
	}

	// TODO check mask bits against the inode.

done:
	if (inode) {
		deref_inode(inode);
	}

	return res;
}

static int graph_unlink(const char *path)
{
	int res = 0;
	struct inode *inode = NULL;

	inode = ref_inode(path, false, 0);	
	if (!inode) {
		res = -errno;
		goto done;
	}

	delete_inode(inode);

done:
	if (inode) {
		deref_inode(inode);
	}

	return res;
}

static int graph_rmdir(const char *path)
{
	int res = 0;
	struct inode *inode = NULL;

	inode = ref_inode(path, false, 0);	
	if (!inode) {
		res = -errno;
		goto done;
	}

	if (!inode->mode & S_IFDIR) {
		errno = ENOTDIR;
		res = -errno;
		goto done;
	}

	if (inode->child != NULL) {
		errno = ENOTEMPTY;
		res = -errno;
		goto done;
	}

	delete_inode(inode);

done:
	if (inode) {
		deref_inode(inode);
	}

	return res;
}

static int graph_rename(const char *from, const char *to)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

static int graph_chmod(const char *path, mode_t mode)
{
	int res = 0;
	struct inode *inode = NULL;

	inode = ref_inode(path, false, 0);	
	if (!inode) {
		res = -errno;
		goto done;
	}

	inode->mode = mode;

done:
	if (inode) {
		deref_inode(inode);
	}

	return res;
}

static int graph_chown(const char *path, uid_t uid, gid_t gid)
{
	int res = 0;
	struct inode *inode = NULL;

	inode = ref_inode(path, false, 0);	
	if (!inode) {
		res = -errno;
		goto done;
	}

	inode->gid = gid;
	inode->uid = uid;

done:
	if (inode) {
		deref_inode(inode);
	}

	return res;
}

static int graph_truncate(const char *path, off_t size)
{
	int res = 0;
	struct inode *inode = NULL;

	inode = ref_inode(path, false, 0);	
	if (!inode) {
		res = -errno;
		goto done;
	}

	if (inode->mode & S_IFDIR) {
		errno = EISDIR;
		res = -EISDIR;
		goto done;
	}

	ftruncate(fileno(inode->f), size);
	inode->size = size;

done:
	if (inode) {
		deref_inode(inode);
	}

	return res;
}

static int graph_utimens(const char *path, const struct timespec ts[2])
{
	 int res = 0;
	 struct inode *inode = NULL;

	 inode = ref_inode(path, false, 0);
	 if (!inode) {
		  res = -errno;
		  goto done;
	 }

	inode->atime = (time_t)ts[0].tv_sec;
	inode->mtime = (time_t)ts[1].tv_sec;

done:
	 if (inode) {
		  deref_inode(inode);
	 }

	 return res;
}

static int graph_open(const char *path, struct fuse_file_info *fi)
{
	 int res = 0;
	 struct inode *inode = NULL;

	 inode = ref_inode(path, (fi->flags & O_CREAT ? true : false), 0777 & S_IFREG);
	 if (!inode) {
		  res = -errno;
		  goto done;
	 }

done:
	 if (inode) {
		  deref_inode(inode);
	 }

	 return res;
}

static int graph_create(const char *path, mode_t mode, struct fuse_file_info *fi)
{
	 int res = 0;
	 struct inode *inode = NULL;

	 inode = ref_inode(path, (fi->flags & O_CREAT ? true : false), mode & S_IFREG);
	 if (!inode) {
		  res = -errno;
		  goto done;
	 }

done:
	 if (inode) {
		  deref_inode(inode);
	 }

	 return res;
}

static int graph_mkdir(const char *path, mode_t mode)
{
	 int res = 0;
	 struct inode *inode = NULL;

	 inode = ref_inode(path, true, mode & S_IFDIR);
	 if (!inode) {
		  res = -errno;
		  goto done;
	 }

done:
	 if (inode) {
		  deref_inode(inode);
	 }

	 return res;
}

static int graph_mknod(const char *path, mode_t mode, dev_t rdev)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

static int graph_fgetattr(const char *path, struct stat *stbuf,
		struct fuse_file_info *fi)
{
	return graph_getattr(path, stbuf);
}

static int graph_ftruncate(const char *path, off_t size,
	struct fuse_file_info *fi)
{
	return graph_truncate(path, size);
}

static int graph_read(const char *path, char *buf, size_t size, off_t offset,
		struct fuse_file_info *fi)
{
    int res = 0;
    struct inode *inode = NULL;

    inode = ref_inode(path, false, 0);
    if (!inode) {
        res = -errno;
        goto done;
    }
    
    if (inode->mode & S_IFDIR) {
        errno = EISDIR;
        res = -EISDIR;
        goto done;
    }

	res = pread(fileno(inode->f), buf, size, offset);
	if (res == -1) {
		res = -errno;
	}

done:
    if (inode) {
        deref_inode(inode);
    }

    return res;
}

static int graph_write(const char *path, const char *buf, size_t size,
		off_t offset, struct fuse_file_info *fi)
{
    int res = 0;
    struct inode *inode = NULL;

    inode = ref_inode(path, false, 0);
    if (!inode) {
        res = -errno;
        goto done;
    }

    if (inode->mode & S_IFDIR) {
        errno = EISDIR;
        res = -EISDIR;
        goto done;
    }

    res = pwrite(fileno(inode->f), buf, size, offset);
    if (res == -1) {
        res = -errno;
    }

done:
    if (inode) {
        deref_inode(inode);
    }

    return res;
}

static int graph_statfs(const char *path, struct statvfs *stbuf)
{
	int res = 0;

	res = statvfs("/", stbuf);
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

	return 0;
}

static int graph_fsync(const char *path, int isdatasync,
		struct fuse_file_info *fi)
{
    int res = 0;
    struct inode *inode = NULL;

    inode = ref_inode(path, false, 0);
    if (!inode) {
        res = -errno;
        goto done;
    }
    
    if (inode->mode & S_IFDIR) {
        errno = EISDIR;
        res = -EISDIR;
        goto done;
    }

    fsync(fileno(inode->f));

done:
    if (inode) {
        deref_inode(inode);
    }

    return res;
}

static int graph_readlink(const char *path, char *buf, size_t size)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

static int graph_symlink(const char *from, const char *to)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

static int graph_link(const char *from, const char *to)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

#ifdef HAVE_SETXATTR
/* xattr operations are optional and can safely be left unimplemented */
static int graph_setxattr(const char *path, const char *name, const char *value,
		size_t size, int flags)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

static int graph_getxattr(const char *path, const char *name, char *value,
		size_t size)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

static int graph_listxattr(const char *path, char *list, size_t size)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
}

static int graph_removexattr(const char *path, const char *name)
{
	// XXX TODO
	errno = EINVAL;
	return -EINVAL;
	int res = 0;
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

int start_unionfs(char *mount_path)
{
	char *argv[4];
	int i;

	init_layers();

	umask(0);

	argv[0] = "graph-unionfs";
	argv[1] = mount_path;
	argv[2] = "-f";

	return fuse_main(3, argv, &graph_oper, NULL);
}

int alloc_unionfs(char *id)
{
	return set_upper(id);
}

int release_unionfs(char *id)
{
	return unset_upper(id);
}

#ifdef STANDALONE_
int main()
{
	start_unionfs("/var/lib/openstorage/fuse");
}
#endif

#endif
