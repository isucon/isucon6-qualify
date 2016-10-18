#!/usr/bin/env perl
use 5.014;
use warnings;
use utf8;
use autodie;

=head1 DESCRIPTION

バグりそうなキーワードを消してしまう

=cut

use Path::Tiny qw/path/;
use JSON::XS qw/decode_json/;

my @skip_words = (
    "1977",
    "PRE",
    "ALAS",
    "バイオダイナミック農法",
    "1010",
    "ポリキューブ",
    "ATOMICA",
    "スターバック",
    "シンドラーグループ",
    "陶都信用農業協同組合",
    "中部地方の道路一覧",
    "中部地方整備局",
    "信濃村",
    "北魚沼農業協同組合",
    "MARK IS",
    "基幹放送事業者",
    "RATO",
    "全国水産・海洋高等学校カッターレース大会",
    "広島市中小企業会館",
    "香川用水記念公園",
    "パームスプリングス・エリアル・トラムウェイ",
    "2015年韓国におけるMERSの流行",
    "兵庫県の観光地",
    "ザ・ビューティーズ!",
    "岩手県の観光地",
    "2728",
    "PRECIOUS",
    "あいおいニッセイ同和損害保険",
    "エルキ",
    "おれんじホープ",
    "AMA",
    "アンダース",
);
my %skip = map {($_ => 1)} @skip_words;

my $f = 'bench/data/ok.json';
my @keywords = do {path($f)->lines};
eval {unlink $f};
open my $fh, '>>', $f;
for my $l (@keywords) {
    my $d = decode_json $l;
    next if $skip{$d->{k}};
    print $fh $l;
}
