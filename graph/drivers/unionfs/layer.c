#ifdef EXPERIMENTAL_

static hashtable_t *layer_hash;

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

// Locate an inode given a path.
static inode *locate(const char *path, bool create)
{
	struct layer *layer = NULL;
	struct stat st;
	char file[PATH_MAX];
	char *fixed_path = NULL;
	char *r = NULL;
	char *dir;
	int base_layer = -1;
	int ret;
	int i;

	layer = get_layer(path, &fixed_path);
	if (!layer) {
		errno = ENOENT;
		goto done;
	}

	strncpy(file, fixed_path, sizeof(file));
	dir = dirname(file);

	for (i = 0; layer; i++) {
		// See if this layer has 'path'

		asprintf(&r, "%s%s", ufs->layers[i], fixed_path);
		if (!r) {
			errno = ENOMEM;
			fprintf(stderr, "Warning, cannot allocate memory\n");
			goto done;
		}

		ret = lstat(r, &st);
		if (ret == 0) {
			// Found the file.

			errno = 0;
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
	if (!r && create) {
		if (base_layer == -1) {
			errno = ENOENT;
			fprintf(stderr, "Warning, create mode requested on %s, but no layer "
					"could be found that could create this file\n", fixed_path);
		} else {
			errno = 0;
			asprintf(&r, "%s%s", ufs->layers[base_layer], fixed_path);
			if (!r) {
				fprintf(stderr, "Warning, cannot allocate memory\n");
				errno = ENOMEM;
			}
		}
	} else {
		errno = ENOENT;
	}

done:

	if (ufs) {
		unlock_ufs(ufs);
	}

	return r;
}

static int insert_inode(struct inode *inode, struct inode *parent,
	const char *name, struct layer *layer)
{
	int ret = 0;
	char *dupname;
	char *base;

	memset(inode, 0, sizeof(struct inode));

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
	inode->mode = S_IFDIR;

	inode->f = tmpfile();
	if (!inode->f) {
		ret = -1;
		goto done;
	}

	pthread_mutex_init(&inode->lock);

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

done:
	if (dupname) {
		free(dupname);
	}

	if (ret) {
		if (inode->f) {
			fclose(inode->f);
		}

		if (inode->name) {
			free(inode->name);
		}
	}

	return ret;
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

	if (insert_inode(&layer->root, NULL, "/", layer)) {
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
