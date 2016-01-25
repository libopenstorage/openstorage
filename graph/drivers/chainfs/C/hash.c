#define _XOPEN_SOURCE 500 /* Enable certain library functions (strdup) on linux.  See feature_test_macros(7) */

#include <stdlib.h>
#include <stdio.h>
#include <limits.h>
#include <string.h>

#include "hash.h"

/* Create a new hashtable. */
hashtable_t *ht_create(int size)
{
	hashtable_t *hashtable = NULL;
	int i;

	if (size < 1) return NULL;

	/* Allocate the table itself. */
	if ((hashtable = malloc(sizeof(hashtable_t))) == NULL) {
		return NULL;
	}

	/* Allocate pointers to the head nodes. */
	if ((hashtable->table = malloc(sizeof(entry_t *) * size)) == NULL) {
		return NULL;
	}

	for(i = 0; i < size; i++) {
		hashtable->table[i] = NULL;
	}

	hashtable->size = size;

	return hashtable;
}

/* Hash a string for a particular hash table. */
int ht_hash(hashtable_t *hashtable, char *key)
{
	unsigned long int hashval = 0;
	int i = 0;

	/* Convert our string to an integer */
	while (hashval < ULONG_MAX && i < strlen(key)) {
		hashval = hashval << 8;
		hashval += key[ i ];
		i++;
	}

	return hashval % hashtable->size;
}

/* Create a key-value pair. */
entry_t *ht_newpair(char *key, void *value)
{
	entry_t *newpair;

	if ((newpair = malloc(sizeof(entry_t))) == NULL) {
		return NULL;
	}

	if ((newpair->key = strdup(key)) == NULL) {
		return NULL;
	}

	newpair->value = value;

	newpair->next = NULL;

	return newpair;
}

/* Insert a key-value pair into a hash table. */
void ht_set(hashtable_t *hashtable, char *key, void *value)
{
	int bin = 0;
	entry_t *newpair = NULL;
	entry_t *next = NULL;
	entry_t *last = NULL;

	bin = ht_hash(hashtable, key);

	next = hashtable->table[bin];

	while (next != NULL && next->key != NULL && strcmp(key, next->key) > 0) {
		last = next;
		next = next->next;
	}

	if (next != NULL && next->key != NULL && strcmp(key, next->key) == 0) {
		/* There's already a pair.  Let's replace that value. */
		next->value = value;
	} else {
		/* Nope, could't find it.  Time to grow a pair. */
		newpair = ht_newpair(key, value);

		if (next == hashtable->table[bin]) {
			/* We're at the start of the linked list in this bin. */
			newpair->next = next;
			hashtable->table[bin] = newpair;
		} else if (next == NULL) {
			/* We're at the end of the linked list in this bin. */
			last->next = newpair;
		} else  {
			/* We're in the middle of the list. */
			newpair->next = next;
			last->next = newpair;
		}
	}
}

/* Retrieve a key-value pair from a hash table. */
void *ht_get(hashtable_t *hashtable, char *key)
{
	int bin = 0;
	entry_t *pair;

	bin = ht_hash(hashtable, key);

	/* Step through the bin, looking for our value. */
	pair = hashtable->table[bin];
	while (pair != NULL && pair->key != NULL && strcmp(key, pair->key) > 0) {
		pair = pair->next;
	}

	/* Did we actually find anything? */
	if (pair == NULL || pair->key == NULL || strcmp(key, pair->key) != 0) {
		return NULL;
	} else {
		return pair->value;
	}
}

void ht_remove(hashtable_t *hashtable, char *key)
{
    int bin = 0;
    entry_t *prev = NULL, *pair;

    bin = ht_hash(hashtable, key);

    /* Step through the bin, looking for our value. */
    pair = hashtable->table[bin];
    while (pair != NULL && pair->key != NULL && strcmp(key, pair->key) > 0) {
		prev = pair;
        pair = pair->next;
    }

    /* Did we actually find anything? */
    if (pair == NULL || pair->key == NULL || strcmp(key, pair->key) != 0) {
        return;
    } else {
		if (prev) {
			prev->next = pair->next;
		}

		if (pair == hashtable->table[bin]) {
			hashtable->table[bin] = pair->next;
		}

		free(pair->key);
		free(pair);
    }
}
