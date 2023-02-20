# SMF-tool

[![GoDoc Reference](https://godoc.org/github.com/majohn-r/SMF-tool?status.svg)](https://pkg.go.dev/github.com/majohn-r/SMF-tool)
[![go.mod](https://img.shields.io/github/go-mod/go-version/majohn-r/SMF-tool)](go.mod)
[![LICENSE](https://img.shields.io/github/license/majohn-r/SMF-tool)](LICENSE)

[![Release](https://img.shields.io/github/v/release/majohn-r/SMF-tool?include_prereleases)](https://github.com/majohn-r/SMF-tool/releases)
[![Code Coverage Report](https://codecov.io/github/majohn-r/SMF-tool/branch/main/graph/badge.svg)](https://codecov.io/github/majohn-r/SMF-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/majohn-r/SMF-tool)](https://goreportcard.com/report/github.com/majohn-r/SMF-tool)
[![Build Status](https://img.shields.io/github/actions/workflow/status/majohn-r/SMF-tool/build.yml?branch=main)](https://github.com/majohn-r/SMF-tool/actions?query=workflow%3Abuild+branch%3Amain)

Command line tool for reading/writing SMF (standard MIDI files).

Very helpful sites for understanding MIDI messages:

* <https://www.midi.org/specifications-old/item/table-1-summary-of-midi-message>
* <https://www.recordingblogs.com/wiki/midi-meta-messages>
* <http://midi.teragonaudio.com/tech/midifile/example.htm>
* <http://www.ccarh.org/courses/253/handout/smf/>
* <https://mido.readthedocs.io/en/latest/meta_message_types.html>

Translating dynamics to MIDI velocity per <https://professionalcomposers.com/music-dynamics-chart/>:

* Pianississimo (ppp) = 16
* Pianissimo (pp) = 33
* Piano (p) = 49
* Mezzo-piano (mp) = 64
* Mezzo-forte (mf) = 80
* Forte (f) = 96
* Fortissimo (ff) = 112
* Fortississimo (fff) = 127

Key signatures <https://www.merriammusic.com/school-of-music/piano-lessons/music-key-signatures/>

use minor variants for 0-7 flats, major variants for 0-7 sharps

* K7b     = KCbMaj = KAbMin  - Ab Bb Cb Db Eb Fb Gb
* K6b     = KGbMaj = KEbMin  - Ab Bb Cb Db Eb    Gb
* K5b     = KDbMaj = KBbMin  - Ab Bb    Db Eb    Gb
* K4b     = KAbMaj = KFMin   - Ab Bb    Db Eb
* K3b     = KEbMaj = KCMin   - Ab Bb       Eb
* K2b     = KBbMaj = KGMin   -    Bb       Eb
* K1b     = KFMaj  = KDMin   -    Bb
* K0b/k0# = KCMaj  = KAMin   - no flats/sharps
* K1#     = KGMaj  = KEMin   -                F#
* K2#     = KDMaj  = KBMin   -       C#       F#
* K3#     = KAMaj  = KF#Min  -       C#       F# G#
* K4#     = KEMaj  = KC#Min  -       C# D#    F# G#
* K5#     = KBMaj  = KG#Min  - A#    C# D#    F# G#
* K6#     = KF#Maj = KD#Min  - A#    C# D# E# F# G#
* K7#     = KC#Maj = KA#Min  - A# B# C# D# E# F# G#

Note values in key signature

* 0 C
* 1 C#/Db
* 2 D
* 3 D#/Eb
* 4 E
* 5 F
* 6 F#/Gb
* 7 G
* 8 G#/Ab
* 9 A
* 10 Bb
* 11 B

BPM values (from JFugue/Staccato)

* GRAVE 40
* LARGO 45
* LARGHETTO 50
* LENTO 55
* ADAGIO 60
* ADAGIETTO 65
* ANDANTE 70
* ADANTINO 80
* MODERATO 95
* ALLEGRETTO 110
* ALLEGRO 120
* VIVACE 145
* PRESTO 180
* PRETISSIMO 220

Predefined instruments (program change message: these are program values)

* Piano
  * 0 PIANO
  * 1 BRIGHT_ACOUSTIC
  * 2 ELECTRIC_GRAND
  * 3 HONKEY_TONK
  * 4 ELECTRIC_PIANO
  * 5 ELECTRIC_PIANO_2
  * 6 HARPSICHORD
  * 7 CLAVINET
* Chromatic Percussion
  * 8 CELESTA
  * 9 GLOCKENSPIEL
  * 10 MUSIC_BOX
  * 11 VIBRAPHONE
  * 12 MARIMBA
  * 13 XYLOPHONE
  * 14 TUBULAR_BELLS
  * 15 DULCIMER
* Organ
  * 16 DRAWBAR_ORGAN
  * 17 PERCUSSIVE_ORGAN
  * 18 ROCK_ORGAN
  * 19 CHURCH_ORGAN
  * 20 REED_ORGAN
  * 21 ACCORDIAN
  * 22 HARMONICA
  * 23 TANGO_ACCORDIAN
* Guitar
  * 24 GUITAR
  * 25 STEEL_STRING_GUITAR
  * 26 ELECTRIC_JAZZ_GUITAR
  * 27 ELECTRIC_CLEAN_GUITAR
  * 28 ELECTRIC_MUTED_GUITAR
  * 29 OVERDRIVEN_GUITAR
  * 30 DISTORTION_GUITAR
  * 31 GUITAR_HARMONICS
* Bass
  * 32 ACOUSTIC_BASS
  * 33 ELECTRIC_BASS_FINGER
  * 34 ELECTRIC_BASS_PICK
  * 35 FRETLESS_BASS
  * 36 SLAP_BASS_1
  * 37 SLAP_BASS_2
  * 38 SYNTH_BASS_1
  * 39 SYNTH_BASS_2
* Strings
  * 40 VIOLIN
  * 41 VIOLA
  * 42 CELLO
  * 43 CONTRABASS
  * 44 TREMOLO_STRINGS
  * 45 PIZZICATO_STRINGS
  * 46 ORCHESTRAL_STRINGS
  * 47 TIMPANI
* Ensemble
  * 48 STRING_ENSEMBLE_1
  * 49 STRING_ENSEMBLE_2
  * 50 SYNTH_STRINGS_1
  * 51 SYNTH_STRINGS_2
  * 52 CHOIR_AAHS
  * 53 VOICE_OOHS
  * 54 SYNTH_VOICE
  * 55 ORCHESTRA_HIT
* Brass
  * 56 TRUMPET
  * 57 TROMBONE
  * 58 TUBA
  * 59 MUTED_TRUMPET
  * 60 FRENCH_HORN
  * 61 BRASS_SECTION
  * 62 SYNTH_BRASS_1
  * 63 SYNTH_BRASS_2
* Reed
  * 64 SOPRANO_SAX
  * 65 ALTO_SAX
  * 66 TENOR_SAX
  * 67 BARITONE_SAX
  * 68 OBOE
  * 69 ENGLISH_HORN
  * 70 BASSOON
  * 71 CLARINET
* Pipe
  * 72 PICCOLO
  * 73 FLUTE
  * 74 RECORDER
  * 75 PAN_FLUTE
  * 76 BLOWN_BOTTLE
  * 77 SKAKUHACHI
  * 78 WHISTLE
  * 79 OCARINA
* Synth Lead
  * 80 SQUARE
  * 81 SAWTOOTH
  * 82 CALLIOPE
  * 83 CHIFF
  * 84 CHARANG
  * 85 VOICE
  * 86 FIFTHS
  * 87 BASS_LEAD
* Synth Pad
  * 88 NEW_AGE
  * 89 WARM
  * 90 POLY_SYNTH
  * 91 CHOIR
  * 92 BOWED
  * 93 METALLIC
  * 94 HALO
  * 95 SWEEP
* Synth Effects
  * 96 RAIN
  * 97 SOUNDTRACK
  * 98 CRYSTAL
  * 99 ATMOSPHERE
  * 100 BRIGHTNESS
  * 101 GOBLINS
  * 102 ECHOES
  * 103 SCI_FI
* Ethnic
  * 104 SITAR
  * 105 BANJO
  * 106 SHAMISEN
  * 107 KOTO
  * 108 KALIMBA
  * 109 BAGPIPE
  * 110 FIDDLE
  * 111 SHANAI
* Percussive
  * 112 TINKLE_BELL
  * 113 AGOGO
  * 114 STEEL_DRUMS
  * 115 WOODBLOCK
  * 116 TAIKO_DRUM
  * 117 MELODIC_DRUM
  * 118 SYNTH_DRUM
  * 119 REVERSE_CYMBAL
* Sound Effects
  * 120 GUITAR_FRET_NOISE
  * 121 BREATH_NOISE
  * 122 SEASHORE
  * 123 BIRD_TWEET
  * 124 TELEPHONE_RING
  * 125 HELICOPTER
  * 126 APPLAUSE
  * 127 GUNSHOT
