#ifndef _LAYER_H_
#define _LAYER_H_
#define _FILE_OFFSET_BITS 64

#include <fuse.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <pthread.h>

#include "hash.h"

// Minimal inode structure.
struct inode {
	pthread_mutex_t lock;

	// Reference count.  TODO this should be a read/write lock.
	int ref;

	// Stat buf.
	mode_t mode;
	nlink_t nlink;
	uid_t uid;
	gid_t gid;
	off_t size;
	time_t atime;
	time_t mtime;
	time_t ctime;

	// Full path name.
	char *full_name;

	// Short path name.
	char *name;

	// The filesystem tree.
	struct inode *parent;
	struct inode *child;
	struct inode *next;
	struct inode *prev;

	// The layer this inode belongs to.
	struct layer *layer;

	// XXX This should point to a block device.
	FILE *f;
};

struct layer {
	char id[PATH_MAX];

	// True if this is the top most later.  This flag is used
	// for determining of modified files should go into this layer.
	bool upper;

	// Handy reference to the root inode.
	struct inode *root;

	// namespace to inode mapping.
	hashtable_t *namespace;

	// Linkage to parent and sibling layers.
	struct layer *parent;
	struct layer *next;
	struct layer *prev;
};

// Create a layer and link it to a parent.  Parent can be "" or NULL.
extern int create_layer(char *id, char *parent_id);

// Remove a layer and all the inodes in this layer.
extern int remove_layer(char *id);

// Returns true if layer exists.
extern int check_layer(char *id);

// Locate an inode given a path.  If 'follow' is specified, then search
// all linked layers for the path.  Create one if 'create' flag is specified.
// Increment reference count on the returned inode.
extern struct inode *ref_inode(const char *path, bool follow,
		bool create, mode_t mode);

// Decrement ref count on an inode.  A node with a refcount of 0 can be 
// deleted if delete_inode is called on it.
extern void deref_inode(struct inode *inode);

// Get statbuf on an inode.  Must be called with reference held.
extern int stat_inode(struct inode *inode, struct stat *stbuf);

// Set mode on an inode.  Must be called with reference held.
extern int chmod_inode(struct inode *inode, mode_t mode);

// Must be called with reference held.
extern int delete_inode(struct inode *inode);

// Rename an inode.  Must be called with reference held.  Returns the inoe
// of the new file.  New inode ref is 1.
extern struct inode *rename_inode(struct inode *inode, const char *to);

// Mark a layer as the top most layer.
extern int set_upper(char *id);

// Unmark a layer as the top most layer.
extern int unset_upper(char *id);

// Fill buf with the dir entries of the root FS.
extern int root_fill(fuse_fill_dir_t filler, char *buf);

// Initialize the layer file system.
extern int init_layers(void);

#endif // _LAYER_H_
