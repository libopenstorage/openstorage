#ifdef EXPERIMENTAL_

int create_layer(char *layer, char *parent)
{
	struct inode *p_ino = NULL;
	struct inode *ino = NULL;
	char *str = NULL;
	int ret = 0;

	ino = ht_get(inode_hash, layer);
	if (ino) {
		ino = NULL;		errno = EEXIST;
		ret = -errno;
		goto done;
	}

	if (parent && parent != "") {
		p_ino = ht_get(inode_hash, parent);
		if (!ino) {
			fprintf(stderr, "Warning, cannot find parent layer %s.\n", parent);
			errno = ENOENT;
			ret = -errno;
			goto done;
		}
	}

	ino = calloc(1, sizeof(struct inode));
	if (!ino) {
		ret = -errno;
		goto done;
	}

	ino->parent = p_ino;
	ino->atime = ino->mtime = ino->ctime = time(NULL);
	ino->uid = getuid();
	ino->gid = getgid();
	ino->mode = S_IFDIR;

	ino->parent = p_ino;
	ino->layer = true;

	pthread_mutex_init(&ino->lock, NULL);

	pthread_mutex_lock(&inode_lock);
	{
		ht_set(inode_hash, layer, ino);
	}
	pthread_mutex_unlock(&inode_lock);

done:
	if (str) {
		free(str);
	}

	if (ret) {
		if (ino) {
			free(ino);
		}

		if (p_ino) {
			free(p_ino);
		}
	}

	return ret;
}

int remove_layer(char *layer)
{
	struct inode *ino = NULL;

	// XXX FIXME
	pthread_mutex_lock(&inode_lock);
	{
		ino = ht_get(inode_hash, layer);
		if (ino) {
			ht_remove(inode_hash, layer);
			free(ino);
		} 
	}
	pthread_mutex_unlock(&inode_lock);
	return 0;
}

#endif // EXPERIMENTAL_
