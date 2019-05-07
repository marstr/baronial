#!/usr/bin/env bash

###############################################################################
# This script finds the unique identifier for the current SCM revision.       #
#                                                                             #
# Input: None                                                                 #
# Output: A single line of text of the format:                                #
#   {commit ID}[-modified]                                                    #
#                                                                             #
# Note:                                                                       #
# If the current state of the local repository does not match the latest      #
# commit, the suffix "-modified" is added"                                    #
###############################################################################

revision="$(git rev-parse HEAD)"

if ! [[ -z "$(git status --short)" ]]; then
	revision="${revision}-modified"
fi

echo "${revision}"
