#!/usr/bin/env bash

###############################################################################
# This script finds the SemVer that would be most applicable to this code,    #
# were it to be distributed in this state.                                    #
#                                                                             #
# Input: None                                                                 #
# Output: A single line of text of the format:                                #
#   {commit ID}[-modified]                                                    #
#                                                                             #
# Note:                                                                       #
# The presence of Git tags will change the behavior of this script. While     #
# this makes it arguably stateful, the benefit is that there need be no       #
# hard-coded version number anywhere in the code-base. This opens the doorway #
# for promotion of a particular commit from a release-candidate to a formally #
# accepted release. However, there are reasons to embed a hard-coded version  #
# into a build. Largely, this is because we do not want to ship the entire    #
# Git repository as part of this application.                                 #
###############################################################################

version=$(./ancestors.sh | ./max-version.pl)
revision=$(./get-revision.sh)


if [[ $(git rev-parse ${version}) != ${revision} ]]; then
    version="${version}-modified"
fi

echo ${version}
