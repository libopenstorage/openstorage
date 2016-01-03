#ifndef _LAYER_H_
#define _LAYER_H_

#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <pthread.h>

#include "hash.h"

// Minimal inode structure.
struct inode {
	int ref;

	mode_t mode;
	nlink_t nlink;
	uid_t uid;
	gid_t gid;
	off_t size;
	time_t atime;
	time_t mtime;
	time_t ctime;

	char *name;

	// The filesystem tree.
	struct inode *parent;
	struct inode *child;
	struct inode *next;

	// The layer this inode belongs to.
	struct layer *layer;

	// XXX This should point to a block device.
	FILE *f;

	// If this flag is set, the inode will be garbage collected.
	bool deleted;

	pthread_mutex_t lock;
};

struct layer {
	char id[256];

	struct inode *root;
	hashtable_t *children;

	struct layer *parent;
};

// Create a layer and link it to a parent.  Parent can be "".
extern int create_layer(char *id, char *parent_id);

// Remove a layer and all the inodes in this layer.
extern int remove_layer(char *id);

// Allocate an inode, add it to the layer and link it to the namespace.
// Initial reference is 1.
extern struct inode *alloc_inode(struct inode *parent, char *name, 
	mode_t mode, struct layer *layer);

// Locate an inode given a path.  Create one if 'create' flag is specified.
// Increment reference count on the inode.
extern struct inode *ref_inode(const char *path, bool create, mode_t mode);

// Decrement ref count on an inode.  A deleted inode with a ref count of 0 
// will be garbage collected.
extern void deref_inode(struct inode *inode);

extern int init_layers(void);

#endif // _LAYER_H_
