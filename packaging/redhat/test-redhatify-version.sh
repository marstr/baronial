#!/usr/bin/env bash

###############################################################################
# This script executes `redhatify-version.sh` against well-known test cases   #
# to ensure consistent and desired behavior.                                  #
#                                                                             #
# Input: None                                                                 #
# Output: The difference between the expected and actual results.             #
###############################################################################

set -e
output_location=$(mktemp -t redhat_versions_XXXXXXXXXX.txt)
cat ./testdata/semvers.txt | perl ./redhatify-version.pl > ${output_location}
diff ./testdata/redhat-vers.txt ${output_location}
rm ${output_location}
