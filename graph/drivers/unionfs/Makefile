all: test

test:
	gcc unionfs.c layer.c hash.c snappy.c -g3 -DEXPERIMENTAL_ -DFILE_OFFSET_BITS=64 -DSTANDALONE_ -lfuse -lulockmgr -lpthread -o unionfs

opt:
	gcc unionfs.c layer.c hash.c snappy.c -O3 -DEXPERIMENTAL_ -DFILE_OFFSET_BITS=64 -DSTANDALONE_ -lfuse -lulockmgr -lpthread -o unionfs

clean:
	rm -rf *.o unionfs
