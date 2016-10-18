#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;
binmode STDOUT, ':encoding(UTF-8)';

# kinshi.csv
# http://monoroch.net/kinshi/

use Regexp::Assemble;
use Text::CSV;

my $ra = Regexp::Assemble->new;
my $csv = Text::CSV->new({
    binary => 1,
    eol    => "\r\n",
});
open my $fh, '<:encoding(UTF-8)', 'testdata/kinshi.csv' or die $!;
while (my $row = $csv->getline($fh)) {
    $ra->add($row->[0]);
}
say $ra->re;
