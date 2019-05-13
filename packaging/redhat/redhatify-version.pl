#!/usr/bin/env perl

###############################################################################
# This script takes a list of semantic versions and converts each to one that #
# is an acceptable version name to embed into an RPM file name.               #
#                                                                             #
# Pragmatically, the difference is that RPM does not allow the character `-`  #
# to be part of a version, whereas SemVer demands that the `-` be used for    #
# defining a 'tag', and also allows the hyphen to be used inside of a tag.    #
# Looking at popular projects, it seems the convention is to replace the `-`  #
# with a `~`.                                                                 #
#                                                                             #
# Input: A list of strings containing entries conforming to semantic version. #
# Output: A list of strings conforming to RedHat Packaging versioning         #
#   guidelines.                                                               #
###############################################################################

use strict;
use warnings FATAL => 'all';

while(my $row = <STDIN>) {
    if ($row =~ /^[vV]?(?<version>\d+(?:\.\d+){0,2})(?<tag>-\S+)?$/) {
        my $transformed = $+{version};
        if (defined $+{tag}) {
            my $cleanedTag = $+{tag};
            $cleanedTag =~ s/-/~/g;
            $transformed = $transformed . $cleanedTag;
        }

        print($transformed . "\n");
    }
}
