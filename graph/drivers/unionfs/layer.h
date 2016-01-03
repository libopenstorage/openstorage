#ifdef EXPERIMENTAL_

#define _LAYER_H_

#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>

#include "hash.h"
#include "layer.h"

// Minimal inode structure.
struct inode {
	mode_t mode;
	nlink_t nlink;
	uid_t uid;
	gid_t gid;
	off_t size;
	time_t atime;
	time_t mtime;
	time_t ctime;

	// The filesystem tree.
	struct inode *parent;
	hashtable_t *children;

	// The layer this inode belongs to.
	struct layer *layer;
};

struct layer {
	char id[256];

	struct inode *root;

	struct layer *parent;
};

// Create a layer and link it to a parent.  Parent can be "".
extern int create_layer(char *id, char *parent);

// Remove a layer and all the inodes in this layer.
extern int remove_layer(char *id);

#endif // _LAYER_H_

#endif // EXPERIMENTAL_
