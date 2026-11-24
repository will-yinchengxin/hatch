//go:build cgo

#import <Foundation/Foundation.h>
#include <sys/stat.h>
#include <string.h>

static BOOL matchesCache(struct stat *a, struct stat *b) {
    return (S_ISDIR(b->st_mode) || (S_ISREG(b->st_mode) && b->st_nlink == 1)) &&
           a->st_dev == b->st_dev && a->st_ino == b->st_ino &&
           a->st_mode == b->st_mode && a->st_uid == b->st_uid && a->st_flags == b->st_flags &&
           a->st_birthtimespec.tv_sec == b->st_birthtimespec.tv_sec &&
           a->st_birthtimespec.tv_nsec == b->st_birthtimespec.tv_nsec &&
           a->st_ctimespec.tv_sec == b->st_ctimespec.tv_sec &&
           a->st_ctimespec.tv_nsec == b->st_ctimespec.tv_nsec &&
           a->st_mtimespec.tv_sec == b->st_mtimespec.tv_sec &&
           a->st_mtimespec.tv_nsec == b->st_mtimespec.tv_nsec;
}

int hatchTrashStorageCache(const char *path, int fd, char **result) {
    @autoreleasepool {
        struct stat held, named, referenced;
        if (fstat(fd, &held) != 0) {
            return 1;
        }
        NSURL *url = [NSURL fileURLWithFileSystemRepresentation:path isDirectory:S_ISDIR(held.st_mode) relativeToURL:nil];
        NSURL *reference = url.fileReferenceURL;
        const char *resolved = reference.filePathURL.fileSystemRepresentation;
        if (resolved == NULL ||
            lstat(path, &named) != 0 || !matchesCache(&held, &named) ||
            lstat(resolved, &referenced) != 0 || !matchesCache(&held, &referenced)) {
            return 1;
        }

        NSURL *destination = nil;
        NSError *error = nil;
        if (![NSFileManager.defaultManager trashItemAtURL:reference resultingItemURL:&destination error:&error]) {
            const char *message = error.localizedDescription.UTF8String;
            *result = strdup(message ?: "Unknown Trash error");
            return 2;
        }
        const char *actual = destination.fileSystemRepresentation;
        if (actual != NULL) {
            *result = strdup(actual);
        }
        return 0;
    }
}
