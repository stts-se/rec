#!/usr/bin/perl

@uttfiles = glob("utt_files/*.utt");

foreach $uttfile (@uttfiles) {
    $uttfile =~ /^.*\/(.+).utt/;
    $base = $1;
    print "$uttfile\n";
    print "$base\n";
    #exit();
    `mkdir audio_dir/$base`;
    `cp utt_files/$base.utt audio_dir/$base`;
}
