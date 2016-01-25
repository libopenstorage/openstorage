#ifndef _HASH_H_
#define _HASH_H_

struct entry_s {
	char *key;
	void *value;
	struct entry_s *next;
};

typedef struct entry_s entry_t;

struct hashtable_s {
	int size;
	struct entry_s **table;	
};

typedef struct hashtable_s hashtable_t;

extern hashtable_t *ht_create(int size);
extern void ht_set(hashtable_t *hashtable, char *key, void *value);
extern void *ht_get(hashtable_t *hashtable, char *key);

#endif	// _HASH_H_

