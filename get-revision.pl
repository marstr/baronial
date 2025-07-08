#!/usr/bin/env perl

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

use strict;
use warnings;

my $revision = `git rev-parse HEAD`;
$revision =~ s/\s+$//;

if(`git status --short` ne ""){
    $revision = $revision . "-modified";
}

print($revision);
