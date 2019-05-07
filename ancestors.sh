#!/usr/bin/env bash

###############################################################################
# This script enumerates all tags in the current repository that point to a   #
# Git commit that is logically a parent of the currently checked-out commit.  #
#                                                                             #
# Input: None                                                                 #
# Output: A list of Git tag names.                                            #
###############################################################################

for tag in $(git tag) ; do
    if git merge-base --is-ancestor ${tag} HEAD; then
        echo ${tag}
    fi
done