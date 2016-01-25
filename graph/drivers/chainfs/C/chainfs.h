#ifdef EXPERIMENTAL_
#ifndef _UNIONFS_H_
#define _UNIONFS_H_
extern int alloc_unionfs(char *id);
extern int release_unionfs(char *id);
extern int start_unionfs(char *mount_path);
#endif // _UNIONFS_H_
#endif // EXPERIMENTAL_
