#!/usr/bin/env perl

###############################################################################
# This script enumerates all tags in the current repository that point to a   #
# Git commit that is logically a parent of the currently checked-out commit.  #
#                                                                             #
# Input: None                                                                 #
# Output: A list of Git tag names.                                            #
###############################################################################

use strict;
use warnings FATAL => 'all';

open(TAGS, "git tag|");

while(my $tag = <TAGS>){
    $tag =~ s/\s+$//;
    system("git merge-base --is-ancestor ${tag} HEAD");
    if($? == 0){
        print($tag . "\n");
    }
}
