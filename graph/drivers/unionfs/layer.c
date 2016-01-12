// gcc layer.c hash.c -DEXPERIMENTAL_ -DFILE_OFFSET_BITS=64 -lfuse -lulockmgr -lpthread -c

#ifndef EXPERIMENTAL_
#define EXPERIMENTAL_
#endif

#ifdef EXPERIMENTAL_

#define _GNU_SOURCE
#define _FILE_OFFSET_BITS 64
#define FUSE_USE_VERSION 26

#include <semaphore.h>
#include <fuse.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <libgen.h>
#include <stdbool.h>
#include <errno.h>
#include <fcntl.h>
#include <dirent.h>
#include <sys/time.h>
#include <sys/types.h>

#include "hash.h"
#include "layer.h"

// A reference to the root inode.
static struct inode *root_inode;

// All layers in the system.
static struct layer *layer_head = NULL;
static hashtable_t *layer_hash;
static pthread_mutex_t layer_lock;

// Guards against a deleted inode getting free'd if we are about to 
// ref count one of it's children.  TODO make this per layer and not global.
static pthread_rwlock_t namespace_lock;

// Delete an inode if the link count is down to 0.
static int reap_inode(struct inode *inode)
{
	struct inode *linked_inode = NULL;
	bool release = true;
	int ret = 0;

	errno = 0;

	if (inode == root_inode) {
		fprintf(stderr, "Warning, trying to delete root inode.\n");
		errno = EINVAL;
		ret = -errno;
		return ret;
	}

	// We will be removing an inode from the namespace.
	pthread_rwlock_wrlock(&namespace_lock);

	// Locks ref from changing.
	pthread_mutex_lock(&inode->lock);

	// We will be making changes to the parent's children links.
	if (inode->parent) {
		pthread_mutex_lock(&inode->parent->lock);
	}

	if (inode->child != NULL) {
		errno = ENOTEMPTY;
		ret = -errno;
		goto done;
	}

	// refcount must be exactly 1.
	if (inode->ref != 1) {
		errno = EBUSY;
		ret = -errno;
		goto done;
	}

	ht_remove(inode->layer->namespace, inode->full_name);

	inode->nlink--;

	if (inode->nlink) {
		// Someone still links to us, so hide this inode but do not delete it.
		// It will get deleted when all it's links go to zero (that is, when a 
		// dependant inode get's deleted.
		release = false;

		// Set this flag so that a linked inode can reap this inode at
		// a later point (when the linked inode is getting deleted).
		inode->deleted = true;
	}

	// Remove this inode from parent.
	if (inode->prev) {
		inode->prev->next = inode->next;
	}

	if (inode->next) {
		inode->next->prev = inode->prev;
	}

	if (inode->parent && inode->parent->child == inode) {
		inode->parent->child = inode->next;
	}

	if (inode->link) {
		linked_inode = inode->link;
	}

done:

	if (inode->parent) {
		pthread_mutex_unlock(&inode->parent->lock);
	}

	// Unset links in case reap is called on this inode again.  This
	// can happen if a linked inode is reaped.
	inode->prev = NULL;
	inode->next = NULL;
	inode->parent = NULL;
	inode->link = NULL;

	pthread_mutex_unlock(&inode->lock);

	if (ret == 0 && release) {
		if (inode->f) {
			fclose(inode->f);
		}

		free(inode);
	}

	pthread_rwlock_unlock(&namespace_lock);

	// Deref linked inode and reap it too.
	if (linked_inode) {
		pthread_mutex_lock(&linked_inode->lock);
		{
			linked_inode->nlink--;
		}
		pthread_mutex_unlock(&linked_inode->lock);

		if (linked_inode->deleted) {
			ret = reap_inode(linked_inode);
		}
	}

	return ret;
}

// Allocate an inode, add it to the layer and link it to the namespace.
// Initial reference is 1.
// Warning - namespace lock must be held.
static struct inode *__alloc_inode(struct inode *parent, char *name,
	mode_t mode, struct layer *layer)
{
	struct inode *inode = NULL;
	struct fuse_context *fuse_ctx;
	char *dupname = NULL;
	char *base;
	int ret = 0;

	if (parent) {
		pthread_mutex_lock(&parent->lock);
	}

	inode = (struct inode *)calloc(1, sizeof(struct inode));
	if (!inode) {
		ret = -errno;
		goto done;
	}

	dupname = strdup(name);
	if (!dupname) {
		errno = ENOMEM;
		ret = -errno;
		goto done;
	}

	inode->ref = 1;

	base = basename(name);
	if (strlen(base) > 255) {
		errno = ENAMETOOLONG;
		ret = -errno;
		goto done;
	}

	inode->full_name = strdup(name);
	inode->name = strdup(base);
	if (!inode->name) {
		errno = ENOMEM;
		ret = -errno;
		goto done;
	}

	inode->atime = inode->mtime = inode->ctime = time(NULL);
	fuse_ctx = fuse_get_context();
	inode->uid = fuse_ctx->uid;
	inode->gid = fuse_ctx->gid;
	inode->mode = mode;
	inode->nlink = 1;

	if (mode & S_IFREG) {
		// XXX this needs to point to a block device.
		inode->f = tmpfile();
		if (!inode->f) {
			errno = ENOMEM;
			ret = -errno;
			goto done;
		}
		fchmod(fileno(inode->f), mode);
	} else {
		inode->f = NULL;
	}

	pthread_mutex_init(&inode->lock, NULL);

	inode->layer = layer;
	inode->child = NULL;
	inode->next = NULL;
	inode->prev = NULL;

	if (parent) {
		inode->parent = parent;

		// Insert into head.  No need to worry about paren'ts ref count
		// since this function can only be called with the reaper lock held.
		inode->next = parent->child;
		if (parent->child) {
			parent->child->prev = inode;
		}
		parent->child = inode;
	}

	if (layer) {
		ht_set(layer->namespace, name, inode);
	}

done:
	if (parent) {
		pthread_mutex_unlock(&parent->lock);
	}

	if (dupname) {
		free(dupname);
	}

	if (ret) {
		if (inode) {
			if (inode->f) {
				fclose(inode->f);
			}

			if (inode->name) {
				free(inode->name);
			}

			free(inode);
			inode = NULL;
		}
	}

	return inode;
}

// Get's the owner layer given a path.
static struct layer *__get_layer(const char *path, char **new_path)
{
	struct layer *layer = NULL;
	char *p, *layer_id = NULL;
	int i, id;

	*new_path = NULL;

	layer_id = strdup(path + 1);
	if (!layer_id) {
		fprintf(stderr, "Warning, cannot allocate memory.\n");
		goto done;
	}

	p = strchr(layer_id, '/');
	if (p) *p = 0;

	*new_path = strchr(path+1, '/');
	if (!*new_path) {
		// Must be a request for root.
		*new_path = "/";
	}

	layer = ht_get(layer_hash, layer_id);

done:
	if (layer_id) {
		free(layer_id);
	}

	if (!layer) {
		errno = ENOENT;
	}

	return layer;
}

// Locate an inode given a path.  If 'follow' is specified, then search
// all linked layers for the path.  Create one if 'create' flag is specified.
// Increment reference count on the returned inode.
struct inode *ref_inode(const char *path, bool follow, 
		ref_mode_t ref_mode, mode_t mode)
{
	struct layer *layer = NULL;
	struct layer *parent_layer = NULL;
	struct inode *inode = NULL;
	struct inode *parent = NULL;
	char file[PATH_MAX];
	char *fixed_path = NULL;
	char *dir;
	int i;

	if (!strcmp(path, "/")) {
		return root_inode;
	}

	if (ref_mode == REF_CREATE || ref_mode == REF_CREATE_EXCL) {
		// Needed because we may be adding new inodes to the namespace
		// and we don't want duplicate entries added for the same name.
		pthread_rwlock_wrlock(&namespace_lock);
	} else {
		// Needed because we will be looking for inodes in the namespace 
		// and we don't want them to get deleted.
		pthread_rwlock_rdlock(&namespace_lock);
	}

	errno = 0;

	parent_layer = layer = __get_layer(path, &fixed_path);
	if (!layer) {
		goto done;
	}

	strncpy(file, fixed_path, sizeof(file));
	dir = dirname(file);

	while (layer) {
		// See if this layer has 'path'
		inode = ht_get(layer->namespace, fixed_path);
		if (inode) {
			if (ref_mode == REF_CREATE_EXCL) {
				// This inode already exists... do not ref it or create a new one.
				inode = NULL;
				errno = EEXIST;
				goto done;
			}

			pthread_mutex_lock(&inode->lock);
			{
				inode->ref++;
			}
			pthread_mutex_unlock(&inode->lock);
			goto done;
		}

		// See if this layer contains the parent directory.  We give
		// preference to the upper layers.
		if (!parent) {
			parent = ht_get(layer->namespace, dir);
			if (parent) {
				parent_layer = layer;
			}

			// No need to lock  or refcount parent since it is used in the zone
			// protected by namespace_lock.
		}

		if (!follow) {
			break;
		}

		layer = layer->parent;
	}

	// If we did not find the file and create mode was requested, construct
	// a file path in the appropriate layer.
	if (!inode && ref_mode == REF_CREATE || ref_mode == REF_CREATE_EXCL) {
		if (!parent) {
			fprintf(stderr, "Warning, create mode requested on %s, but no layer "
					"could be found that could create this file\n", fixed_path);
		} else {
			// At this point, parent is safe to use because we have the
			// reaper lock held.
			inode = __alloc_inode(parent, fixed_path, mode, parent_layer);
		}
	}

done:

	pthread_rwlock_unlock(&namespace_lock);

	if (!inode) {
		if (!errno) {
			errno = ENOENT;
		}
	}

	return inode;
}

// Decrement ref count on an inode.  A node with a refcount of 0 can be 
// deleted if delete_inode is called on it.
void deref_inode(struct inode *inode)
{
	errno = 0;

	if (inode == root_inode) {
		return;
	}

	pthread_mutex_lock(&inode->lock);
	{
		inode->ref--;
	}
	pthread_mutex_unlock(&inode->lock);
}

// Get statbuf on an inode.  Must be called with reference held.
int stat_inode(struct inode *inode, struct stat *stbuf)
{
	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	if (inode->f) {
		fstat(fileno(inode->f), stbuf);
		stbuf->st_nlink = inode->nlink;
	} else {
		stbuf->st_mode = inode->mode;
		stbuf->st_nlink = inode->nlink;
		stbuf->st_uid = inode->uid;
		stbuf->st_gid = inode->gid;
		stbuf->st_size = inode->size;
		stbuf->st_atime = inode->atime;
		stbuf->st_mtime = inode->mtime;
		stbuf->st_ctime = inode->ctime;
		stbuf->st_ino = 0;
	}

	return 0;
}

// Set mode on an inode.  Must be called with reference held.
int chmod_inode(struct inode *inode, mode_t mode)
{
	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	if (inode->f) {
		fchmod(fileno(inode->f), mode);
	} else {
		inode->mode = mode;
	}

	return 0;
}

// Set uid and gid on an inode.  Must be called with reference held.
int chown_inode(struct inode *inode, uid_t uid, gid_t gid)
{
	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	inode->gid = gid;
	inode->uid = uid;

	return 0;
}

// Truncate an inode.  Must be called with reference held.
int truncate_inode(struct inode *inode, off_t size)
{
	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	ftruncate(fileno(inode->f), size);
	inode->size = size;
}

// Read from an inode.  Must be called with reference held.
int read_inode(struct inode *inode, char *buf, size_t size, off_t offset)
{
	int res = 0;

	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	res = pread(fileno(inode->f), buf, size, offset);
	if (res == -1) {
		res = -errno;
	}

	return res;
}

// Write to an inode.  Must be called with reference held.
int write_inode(struct inode *inode, const char *buf, size_t size, off_t offset)
{
	int res = 0;

	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	res = pwrite(fileno(inode->f), buf, size, offset);
	if (res == -1) {
		res = -errno;
	}

	return res;
}

// Sync an inode.  Must be called with reference held.
int sync_inode(struct inode *inode)
{
	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	fsync(fileno(inode->f));

	return 0;
}

// Get the mode of the inode.  Refcount must be held.
mode_t get_inode_mode(struct inode *inode)
{
	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	return inode->mode;
}

// Set atime and mtime on an inode.  Must be called with reference held.
int utimens_inode(struct inode *inode, time_t atime, time_t mtime)
{
	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	inode->atime = atime;
	inode->mtime = mtime;

	return 0;
}

// Create a new namespace link on an inode.  Must be called with reference held.
int link_inode(struct inode *inode, const char *to)
{
	struct inode *new_inode = NULL;
	int ret = 0;

	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	if (inode->mode & S_IFDIR) {
		errno = EPERM;
		ret = -errno;
	}

	new_inode = ref_inode(to, true, REF_CREATE_EXCL, 0);
	if (!new_inode) {
		ret = -errno;
		goto done;
	}

	new_inode->link = inode;

	pthread_mutex_lock(&inode->lock);
	{
		inode->nlink++;
	}
	pthread_mutex_unlock(&inode->lock);

done:

	return ret;
}

// Rename an inode.  Must be called with reference held.
struct inode *rename_inode(struct inode *inode, const char *to)
{
	struct inode *new_inode = NULL;
	struct inode *orig = inode;

	errno = 0;

	if (inode->link) {
		inode = inode->link;
	}

	// XXX FIXME old name and new name are in the name space at the same time
	// for the duration of this function.  Need to atomically swap the namespace.
	// This can be done by leveraging the REF_CREATE_EXCL flag.
	new_inode = ref_inode(to, true, REF_CREATE_EXCL, 0);
	if (!new_inode) {
		goto done;
	}

	// Just make the new node a link to the previous node.
	// XXX FIXME this needs to be done better.
	new_inode->link = inode;
	pthread_mutex_lock(&inode->lock);
	{
		inode->nlink++;
	}
	pthread_mutex_unlock(&inode->lock);

	// Release orig if it was a link itself.
	if (orig != inode) {
		if (reap_inode(orig) != 0) {
			reap_inode(new_inode);
			new_inode = NULL;
			goto done;
		}
	}

done:

	return new_inode;
}

// Must be called with reference held.
int delete_inode(struct inode *inode)
{
	int ret = 0;

	errno = 0;

	if (inode->parent == NULL) {
		// Cannot delete root inodes... Those can only be
		// removed by calling remove_layer.
		errno = EBUSY;
		ret = -errno;
		goto done;
	}

	ret = reap_inode(inode);
	if (ret) {
		goto done;
	}

done:
	return ret;
}

// Create a layer and link it to a parent.  Parent can be "" or NULL.
int create_layer(char *id, char *parent_id)
{
	struct layer *parent = NULL;
	struct layer *layer = NULL;
	char *str = NULL;
	int ret = 0;

	if (!layer_hash) {
		errno = EINVAL;
		ret = -errno;
		goto done;
	}

	layer = ht_get(layer_hash, id);
	if (layer) {
		layer = NULL;
		errno = EEXIST;
		ret = -errno;
		goto done;
	}

	if (parent_id && strcmp(parent_id, "")) {
		parent = ht_get(layer_hash, parent_id);
		if (!parent) {
			fprintf(stderr, "Warning, cannot find parent layer %s.\n", parent_id);
			errno = ENOENT;
			ret = -errno;
			goto done;
		}
	}

	layer = calloc(1, sizeof(struct layer));
	if (!layer) {
		ret = -errno;
		goto done;
	}

	strncpy(layer->id, id, sizeof(layer->id));

	layer->namespace = ht_create(65536);

	// Layer namespace creation.
	layer->root = __alloc_inode(NULL, "/", 0777 | S_IFDIR, layer);
	if (layer->root == NULL) {
		ret = -errno;
		goto done;
	}

	deref_inode(layer->root);

	// Layer linkages.
	layer->upper = false;
	layer->parent = parent;
	pthread_mutex_lock(&layer_lock);
	{
		ht_set(layer_hash, id, layer);
		layer->next = layer_head;
		if (layer_head) {
			layer_head->prev = layer;
		}
		layer_head = layer;
	}
	pthread_mutex_unlock(&layer_lock);

	fprintf(stderr, "Created layer %s\n", id);

done:
	if (str) {
		free(str);
	}

	if (ret) {
		if (layer) {
			free(layer);
		}
	}

	return ret;
}

static int remove_inodes(struct layer *layer)
{
	// TODO

	return 0;
}

int remove_layer(char *id)
{
	int ret = 0;

	return 0;

	pthread_mutex_lock(&layer_lock);
	{
		struct layer *layer = ht_get(layer_hash, id);

		if (layer) {
			ht_remove(layer_hash, id);
			if (layer->next) {
				layer->next->prev = layer->prev;
			}

			if (layer->prev) {
				layer->prev->next = layer->next;
			}

			if (layer == layer_head) {
				layer_head = layer->next;
			}

			if (remove_inodes(layer)) {
				errno = EBUSY;
				ret = -errno;
			}
		} else {
			errno = ENOENT;
			ret = -errno;
		}
	}
	pthread_mutex_unlock(&layer_lock);

	return ret;
}

// Returns true if layer exists.
int check_layer(char *id)
{
	bool ret = false;

	pthread_mutex_lock(&layer_lock);
	{
		struct layer *layer = ht_get(layer_hash, id);
		if (layer) {
			ret = true;
		}
	}
	pthread_mutex_unlock(&layer_lock);

	return ret;
}

// Fill buf with the dir entries of the root FS.
int root_fill(fuse_fill_dir_t filler, char *buf)
{
	int ret = 0;

	pthread_mutex_lock(&layer_lock);
	{
		struct layer *layer = layer_head;

		while (layer) {
			struct stat st;
			char d_name[PATH_MAX];

			snprintf(d_name, sizeof(d_name), "%s", layer->id);

			st.st_mode = layer->root->mode;
			st.st_nlink = layer->root->nlink;
			st.st_uid = layer->root->uid;
			st.st_gid = layer->root->gid;
			st.st_size = layer->root->size;
			st.st_atime = layer->root->atime;
			st.st_mtime = layer->root->mtime;
			st.st_ctime = layer->root->ctime;
			st.st_ino = 0;

			if (filler(buf, d_name, &st, 0)) {
				errno = ENOMEM;
				ret = -errno;

				fprintf(stderr, "Warning, Filler too full on root.\n");
				break;
			}

			layer = layer->next;
		}
	}
	pthread_mutex_unlock(&layer_lock);

	return ret;
}

// Mark a layer as the top most layer.
int set_upper(char *id)
{
	struct layer *layer = NULL;

	layer = ht_get(layer_hash, id);
	if (!layer) {
		errno = ENOENT;
		return -1;
	}

	layer->upper = true;

	errno = 0;
	return 0;
}

// Unmark a layer as the top most layer.
int unset_upper(char *id)
{
	struct layer *layer = NULL;

	layer = ht_get(layer_hash, id);
	if (!layer) {
		errno = ENOENT;
		return -1;
	}

	layer->upper = false;

	errno = 0;
	return 0;
}

int init_layers()
{
	pthread_t tid;

	layer_hash = ht_create(65536);
	if (!layer_hash) {
		return -1;
	}

	root_inode = __alloc_inode(NULL, "/", 0777 | S_IFDIR, NULL);

	pthread_rwlock_init(&namespace_lock, 0);
	pthread_mutex_init(&layer_lock, 0);

	return 0;
}

#endif // EXPERIMENTAL_
