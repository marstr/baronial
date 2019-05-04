#!/usr/bin/env perl

use strict;
use warnings FATAL => 'all';

while(my $row = <STDIN>) {
    if ($row =~ /^[vV]?(?<version>\d+(?:\.\d+){0,2})(?:-(?<tag>\S+))?$/) {
        my $transformed = $+{version};


        if (defined $+{tag}) {
            $transformed = $transformed . "~" . $+{tag};
        }

        print($transformed . "\n");
    }
}
