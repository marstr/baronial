#!/usr/bin/env perl

###############################################################################
# This script evaluates a list containing strings conforming to a SemVer, and #
# finds the entry indicating the most-recent release.                         #
#                                                                             #
# For example:                                                                #
#  1.0.0-beta < 1.0.0 < 1.1 < 1.2.1                                           #
#                                                                             #
# Input: A list containing entries with semantic versions.                    #
# Output: A single line of text conforming to semantic versioning.            #
#                                                                             #
# Note: this script should not be relied on for discerning the max version    #
# between versions that differ only by tag. For instance, the behavior        #
# between 1.0.1-foo and 1.0.1-bar is undefined.                               #
###############################################################################

use strict;
use warnings;

my $maxMajor = 0;
my $maxMinor = 0;
my $maxPatch = 0;
my $maxTag = "";

while(my $row = <STDIN>) {
    if ($row =~ /^[vV]?(?<major>\d+)(?:\.(?<minor>\d+))?(?:\.(?<patch>\d+))?(?:-(?<tag>\S+))?$/) {
        my $currentMajor = int($+{major});
        my $currentMinor = int($+{minor});
        my $currentPatch = int($+{patch});
        my $currentTag = $+{tag};

        if (not defined $currentMinor) {
            $currentMinor = 0;
        }

        if (not defined $currentPatch) {
            $currentPatch = 0;
        }

        if (not defined $currentTag) {
            $currentTag = "";
        }

        if ($currentMajor > $maxMajor) {
            $maxMajor = $currentMajor;
            $maxMinor = $currentMinor;
            $maxPatch = $currentPatch;
            $maxTag = $currentTag;
            next;
        } elsif ($currentMajor < $maxMajor) {
            next;
        }

        if ($currentMinor > $maxMinor) {
            $maxMinor = $currentMinor;
            $maxPatch = $currentPatch;
            $maxTag = $currentTag;
            next;
        } elsif ($currentMinor < $maxMinor) {
            next;
        }

        if ($currentPatch > $maxPatch) {
            $maxPatch = $currentPatch;
            $maxTag = $currentTag;
            next;
        } elsif ($currentPatch < $maxPatch) {
            next;
        }

        if ($currentTag eq "" and $maxTag ne "") {
            $maxTag = $currentTag;
            next;
        } elsif ($maxTag eq "" and $currentTag ne ""){
            next;
        } elsif ($currentTag gt $maxTag) {
            $maxTag = $currentTag;
            next
        } elsif ($currentTag le $maxTag) {
            next;
        }
    }
}

my $formatted = "v${maxMajor}.${maxMinor}.${maxPatch}";
if ($maxTag ne "") {
    $formatted = $formatted . "-${maxTag}";
}

print($formatted . "\n");
