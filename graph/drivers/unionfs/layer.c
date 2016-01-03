// gcc layer.c hash.c -DEXPERIMENTAL_ -DiFILE_OFFSET_BITS=64 -lfuse -lulockmgr -lpthread -c
#ifdef EXPERIMENTAL_

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

static hashtable_t *layer_hash;

static struct inode *alloc_inode(struct inode *parent, char *name, 
	mode_t mode, struct layer *layer)
{
	struct inode *inode;
	int ret = 0;
	char *dupname;
	char *base;

	inode = calloc(1, sizeof(struct inode));
	if (!inode) {
		ret = -1;
		goto done;
	}

	dupname = strdup(name);
	if (!dupname) {
		ret = -1;
		goto done;
	}

	inode->deleted = false;

	base = basename(name);
	inode->name = strdup(base);
	if (!inode->name) {
		ret = -1;
		goto done;
	}

	inode->atime = inode->mtime = inode->ctime = time(NULL);
	inode->uid = getuid();
	inode->gid = getgid();
	inode->mode = mode;

	inode->f = tmpfile();
	if (!inode->f) {
		ret = -1;
		goto done;
	}

	pthread_mutex_init(&inode->lock, NULL);

	inode->layer = layer;

	if (parent) {
		inode->parent = parent;

		pthread_mutex_lock(&parent->lock);
		{
			inode->next = parent->child;
			parent->child = inode;
		}
		pthread_mutex_unlock(&parent->lock);
	}

	ht_set(layer->children, name, inode);

done:
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

static struct layer *get_layer(const char *path, char **new_path)
{
	struct layer *layer = NULL;
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

	layer = ht_get(layer_hash, tmp_path);

done:
	if (tmp_path) {
		free(tmp_path);
	}

	return layer;
}

// Locate an inode given a path.  Create one if 'create' flag is specified.
static struct inode *locate(const char *path, bool create, mode_t mode)
{
	struct layer *layer = NULL;
	struct layer *parent_layer = NULL;
	struct inode *inode = NULL;
	struct inode *parent = NULL;
	char file[PATH_MAX];
	char *fixed_path = NULL;
	char *dir;
	int i;

	errno = 0;

	parent_layer = layer = get_layer(path, &fixed_path);
	if (!layer) {
		goto done;
	}

	strncpy(file, fixed_path, sizeof(file));
	dir = dirname(file);

	while (layer) {
		// See if this layer has 'path'
		inode = ht_get(layer->children, fixed_path);
		if (inode) {
			goto done;
		}

		// See if this layer contains the parent directory.  We give
		// preference to the upper layers.
		if (!parent) {
			parent = ht_get(layer->children, dir);
			parent_layer = layer;
		}

		layer = layer->parent;
	}

	// If we did not find the file and create mode was requested, construct
	// a file path in the appropriate layer.	
	if (!inode && create) {
		if (!parent) {
			fprintf(stderr, "Warning, create mode requested on %s, but no layer "
					"could be found that could create this file\n", fixed_path);
		} else {
			inode = alloc_inode(parent, fixed_path, mode, parent_layer);
		}
	}

done:

	if (!inode) {
		if (!errno) {
			errno = ENOENT;
		}
	}

	return inode;
}

int create_layer(char *id, char *parent_id)
{
	struct layer *parent = NULL;
	struct layer *layer = NULL;
	char *str = NULL;
	int ret = 0;

	layer = ht_get(layer_hash, id);
	if (layer) {
		errno = EEXIST;
		layer = NULL;
		ret = -errno;
		goto done;
	}

	if (parent_id && parent_id != "") {
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

	layer->parent = parent;
	layer->children = ht_create(65536);

	layer->root = alloc_inode(NULL, "/", S_IFDIR, layer);
	if (layer->root == NULL) {
		ret = -errno;
		goto done;
	}

	ht_set(layer_hash, id, layer);

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

int remove_layer(char *id)
{
	ht_remove(layer_hash, id);

	// XXX remove inodes lazily.

	return 0;
}

int init_layers()
{
	layer_hash = ht_create(65536);
	if (!layer_hash) {
		return -1;
	}

	return 0;
}

#endif // EXPERIMENTAL_
