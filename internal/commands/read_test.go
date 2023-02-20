package commands

import (
	"bytes"
	"encoding/binary"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"gitlab.com/gomidi/midi/v2/smf"
)

func Test_newRead(t *testing.T) {
	type args struct {
		c     *tools.Configuration
		flags *flag.FlagSet
	}
	tests := map[string]struct {
		args
		want  tools.CommandProcessor
		want1 bool
		output.WantedRecording
	}{
		"basic": {
			args:  args{c: tools.EmptyConfiguration(), flags: flag.NewFlagSet("read", flag.ContinueOnError)},
			want:  &read{key: &smf.Key{IsMajor: true}},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := newRead(o, tt.args.c, tt.args.flags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newRead() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newRead() got1 = %v, want %v", got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("newRead() %s", issue)
				}
			}
		})
	}
}

func Test_newReadCommand(t *testing.T) {
	type args struct {
		c     *tools.Configuration
		flags *flag.FlagSet
	}
	tests := map[string]struct {
		args
		want  tools.CommandProcessor
		want1 bool
		output.WantedRecording
	}{
		"basic": {
			args:  args{c: tools.EmptyConfiguration(), flags: flag.NewFlagSet("read", flag.ContinueOnError)},
			want:  &read{key: &smf.Key{IsMajor: true}},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := newReadCommand(o, tt.args.c, tt.args.flags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newReadCommand() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newReadCommand() got1 = %v, want %v", got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("newReadCommand() %s", issue)
				}
			}
		})
	}
}

type trackData struct {
	data []byte
}

type eventData struct {
	data []byte
}

func makeMIDITrack(events []eventData) (content trackData) {
	content.data = append(content.data, []byte("MTrk")...)
	var total uint32
	for _, event := range events {
		total += uint32(len(event.data))
	}
	content.data = append(content.data, encode32(total)...)
	for _, event := range events {
		content.data = append(content.data, event.data...)
	}
	return
}

func makeMIDIFileContent(header []byte, tracks []trackData) (content []byte) {
	content = append(content, header...)
	for _, track := range tracks {
		content = append(content, track.data...)
	}
	return
}

func encode32(raw uint32) []byte {
	a := make([]byte, 4)
	binary.BigEndian.PutUint32(a, raw)
	return a
}

func encode16(raw uint16) []byte {
	a := make([]byte, 2)
	binary.BigEndian.PutUint16(a, raw)
	return a
}

func makeMIDIFileHeader(format, tracks, ticks uint16) (content []byte) {
	content = append(content, []byte("MThd")...)
	content = append(content, encode32(6)...)
	content = append(content, encode16(format)...)
	content = append(content, encode16(tracks)...)
	content = append(content, encode16(ticks&0x7FFF)...)
	return
}

// shamelessly borrowed from gitlab.com/gomidi/midi/v2/internal/utils
const (
	vlqContinue = 128
)

func VlqEncode(n uint32) (out []byte) {
	var quo uint32
	quo = n / vlqContinue

	out = append(out, byte(n%vlqContinue))

	for quo > 0 {
		out = append(out, byte(quo)|vlqContinue)
		quo /= vlqContinue
	}

	reverse(out)
	return
}

func reverse(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

// end shamelessly borrowed from gitlab.com/gomidi/midi/v2/internal/utils

func makeEvent(delta uint32, message []byte) eventData {
	eD := eventData{}
	eD.data = append(eD.data, VlqEncode(delta)...)
	eD.data = append(eD.data, message...)
	return eD
}

var metaEndOfTrackMsg = []byte{0xFF, 0x2F, 0}

func makeTrivialContent() []byte {
	header := makeMIDIFileHeader(1, 16, 120)
	trackContent := makeEvent(0, metaEndOfTrackMsg)
	var tracks []trackData
	for k := 0; k < 16; k++ {
		tracks = append(tracks, makeMIDITrack([]eventData{trackContent}))
	}
	return makeMIDIFileContent(header, tracks)
}

func Test_read_Exec(t *testing.T) {
	type args struct {
		args []string
	}
	tests := map[string]struct {
		r        *read
		preTest  func()
		postTest func()
		args
		wantOk bool
		output.WantedRecording
	}{
		"no args": {
			r:        &read{key: &smf.Key{IsMajor: true}},
			preTest:  func() {},
			postTest: func() {},
			wantOk:   false,
			WantedRecording: output.WantedRecording{
				Error: "You disabled all functionality for the command \"read\".\n",
				Log:   "level='error'  msg='the user disabled all functionality'\n",
			},
		},
		"bad arg": {
			r: &read{key: &smf.Key{IsMajor: true}},
			preTest: func() {
				_ = tools.Mkdir("badArgs")
				_ = tools.CreateFile(filepath.Join("badArgs", "file.mid"), []byte{})
			},
			postTest: func() {
				_ = os.RemoveAll("badArgs")
			},
			args:            args{args: []string{filepath.Join("badArgs", "file.mid")}},
			wantOk:          false,
			WantedRecording: output.WantedRecording{Error: "An error occurred while reading \"badArgs\\\\file.mid\": EOF.\n"},
		},
		"good arg": {
			r: &read{key: &smf.Key{IsMajor: true}},
			preTest: func() {
				_ = tools.Mkdir("goodArgs")
				_ = tools.CreateFile(filepath.Join("goodArgs", "file.mid"), makeTrivialContent())
			},
			postTest: func() {
				_ = os.RemoveAll("goodArgs")
			},
			args:   args{args: []string{filepath.Join("goodArgs", "file.mid")}},
			wantOk: true,
			WantedRecording: output.WantedRecording{Console: "" +
				"File: \"goodArgs\\\\file.mid\"\n" +
				"Quarter note: 120 ticks\n" +
				"16 tracks\n" +
				"Track 0 is empty\n" +
				"Track 1 is empty\n" +
				"Track 2 is empty\n" +
				"Track 3 is empty\n" +
				"Track 4 is empty\n" +
				"Track 5 is empty\n" +
				"Track 6 is empty\n" +
				"Track 7 is empty\n" +
				"Track 8 is empty\n" +
				"Track 9 is empty\n" +
				"Track 10 is empty\n" +
				"Track 11 is empty\n" +
				"Track 12 is empty\n" +
				"Track 13 is empty\n" +
				"Track 14 is empty\n" +
				"Track 15 is empty\n" +
				"EOF \"goodArgs\\\\file.mid\"\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			o := output.NewRecorder()
			if gotOk := tt.r.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("read.Exec() = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.Exec() %s", issue)
				}
			}
		})
	}
}

func Test_read_asNote(t *testing.T) {
	type args struct {
		channel uint8
		raw     uint8
	}
	tests := map[string]struct {
		r *read
		args
		want string
	}{
		// C Major
		"major key 0":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 0}, want: "C0"},
		"major key 1":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 1}, want: "C♯0"},
		"major key 2":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 2}, want: "D0"},
		"major key 3":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 3}, want: "D♯0"},
		"major key 4":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 4}, want: "E0"},
		"major key 5":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 5}, want: "F0"},
		"major key 6":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 6}, want: "F♯0"},
		"major key 7":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 7}, want: "G0"},
		"major key 8":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 8}, want: "G♯0"},
		"major key 9":  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 9}, want: "A0"},
		"major key 10": {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 10}, want: "A♯0"},
		"major key 11": {r: &read{key: &smf.Key{IsMajor: true}}, args: args{raw: 11}, want: "B0"},
		// A minor
		"minor key 0":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 0}, want: "C0"},
		"minor key 1":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 1}, want: "D♭0"},
		"minor key 2":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 2}, want: "D0"},
		"minor key 3":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 3}, want: "E♭0"},
		"minor key 4":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 4}, want: "E0"},
		"minor key 5":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 5}, want: "F0"},
		"minor key 6":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 6}, want: "G♭0"},
		"minor key 7":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 7}, want: "G0"},
		"minor key 8":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 8}, want: "A♭0"},
		"minor key 9":  {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 9}, want: "A0"},
		"minor key 10": {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 10}, want: "B♭0"},
		"minor key 11": {r: &read{key: &smf.Key{Key: 9, IsFlat: true}}, args: args{raw: 11}, want: "B0"},
		// percussion
		"unknown 34":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 34}, want: "unknown percussion 34"},
		"ACOUSTIC_BASS_DRUM": {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 35}, want: "ACOUSTIC_BASS_DRUM"},
		"BASS_DRUM":          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 36}, want: "BASS_DRUM"},
		"SIDE_STICK":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 37}, want: "SIDE_STICK"},
		"ACOUSTIC_SNARE":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 38}, want: "ACOUSTIC_SNARE"},
		"HAND_CLAP":          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 39}, want: "HAND_CLAP"},
		"ELECTRIC_SNARE":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 40}, want: "ELECTRIC_SNARE"},
		"LO_FLOOR_TOM":       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 41}, want: "LO_FLOOR_TOM"},
		"CLOSED_HI_HAT":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 42}, want: "CLOSED_HI_HAT"},
		"HIGH_FLOOR_TOM":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 43}, want: "HIGH_FLOOR_TOM"},
		"PEDAL_HI_HAT":       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 44}, want: "PEDAL_HI_HAT"},
		"LO_TOM":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 45}, want: "LO_TOM"},
		"OPEN_HI_HAT":        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 46}, want: "OPEN_HI_HAT"},
		"LO_MID_TOM":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 47}, want: "LO_MID_TOM"},
		"HI_MID_TOM":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 48}, want: "HI_MID_TOM"},
		"CRASH_CYMBAL_1":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 49}, want: "CRASH_CYMBAL_1"},
		"HI_TOM":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 50}, want: "HI_TOM"},
		"RIDE_CYMBAL_1":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 51}, want: "RIDE_CYMBAL_1"},
		"CHINESE_CYMBAL":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 52}, want: "CHINESE_CYMBAL"},
		"RIDE_BELL":          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 53}, want: "RIDE_BELL"},
		"TAMBOURINE":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 54}, want: "TAMBOURINE"},
		"SPLASH_CYMBAL":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 55}, want: "SPLASH_CYMBAL"},
		"COWBELL":            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 56}, want: "COWBELL"},
		"CRASH_CYMBAL_2":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 57}, want: "CRASH_CYMBAL_2"},
		"VIBRASLAP":          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 58}, want: "VIBRASLAP"},
		"RIDE_CYMBAL_2":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 59}, want: "RIDE_CYMBAL_2"},
		"HI_BONGO":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 60}, want: "HI_BONGO"},
		"LO_BONGO":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 61}, want: "LO_BONGO"},
		"MUTE_HI_CONGA":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 62}, want: "MUTE_HI_CONGA"},
		"OPEN_HI_CONGA":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 63}, want: "OPEN_HI_CONGA"},
		"LO_CONGA":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 64}, want: "LO_CONGA"},
		"HI_TIMBALE":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 65}, want: "HI_TIMBALE"},
		"LO_TIMBALE":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 66}, want: "LO_TIMBALE"},
		"HI_AGOGO":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 67}, want: "HI_AGOGO"},
		"LO_AGOGO":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 68}, want: "LO_AGOGO"},
		"CABASA":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 69}, want: "CABASA"},
		"MARACAS":            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 70}, want: "MARACAS"},
		"SHORT_WHISTLE":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 71}, want: "SHORT_WHISTLE"},
		"LONG_WHISTLE":       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 72}, want: "LONG_WHISTLE"},
		"SHORT_GUIRO":        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 73}, want: "SHORT_GUIRO"},
		"LONG_GUIRO":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 74}, want: "LONG_GUIRO"},
		"CLAVES":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 75}, want: "CLAVES"},
		"HI_WOOD_BLOCK":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 76}, want: "HI_WOOD_BLOCK"},
		"LO_WOOD_BLOCK":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 77}, want: "LO_WOOD_BLOCK"},
		"MUTE_CUICA":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 78}, want: "MUTE_CUICA"},
		"OPEN_CUICA":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 79}, want: "OPEN_CUICA"},
		"MUTE_TRIANGLE":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 80}, want: "MUTE_TRIANGLE"},
		"OPEN_TRIANGLE":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 81}, want: "OPEN_TRIANGLE"},
		"unknown 82":         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, raw: 82}, want: "unknown percussion 82"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.r.asNote(tt.args.channel, tt.args.raw); got != tt.want {
				t.Errorf("read.asNote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_read_asVolume(t *testing.T) {
	type args struct {
		velocity uint8
	}
	tests := map[string]struct {
		r *read
		args
		want string
	}{
		"very quiet":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 8}, want: "below pianississimo (𝆏𝆏𝆏) (8)"},
		"pianississimo":   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 16}, want: "pianississimo (𝆏𝆏𝆏)"},
		"a little louder": {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 24}, want: "between pianississimo (𝆏𝆏𝆏) and pianissimo (𝆏𝆏) (24)"},
		"pianissimo":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 33}, want: "pianissimo (𝆏𝆏)"},
		"even louder":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 40}, want: "between pianissimo (𝆏𝆏) and piano (𝆏) (40)"},
		"piano":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 49}, want: "piano (𝆏)"},
		"bit more":        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 56}, want: "between piano (𝆏) and mezzo-piano (𝆐𝆏) (56)"},
		"mezzo-piano":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 64}, want: "mezzo-piano (𝆐𝆏)"},
		"hear it?":        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 72}, want: "between mezzo-piano (𝆐𝆏) and mezzo-forte (𝆐𝆑) (72)"},
		"mezzo-forte":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 80}, want: "mezzo-forte (𝆐𝆑)"},
		"loud":            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 88}, want: "between mezzo-forte (𝆐𝆑) and forte (𝆑) (88)"},
		"forte":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 96}, want: "forte (𝆑)"},
		"quite loud":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 104}, want: "between forte (𝆑) and fortissimo (𝆑𝆑) (104)"},
		"fortissimo":      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 112}, want: "fortissimo (𝆑𝆑)"},
		"LOUD":            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 120}, want: "between fortissimo (𝆑𝆑) and fortississimo (𝆑𝆑𝆑) (120)"},
		"fortississimo":   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 127}, want: "fortississimo (𝆑𝆑𝆑)"},
		"deafening":       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{velocity: 128}, want: "above fortississimo (𝆑𝆑𝆑) (128)"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.r.asVolume(tt.args.velocity); got != tt.want {
				t.Errorf("read.asVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeAfterTouchMessage(channel, pressure int) smf.Message {
	ch := byte(channel & 0x0F)
	p := byte(pressure & 0x07F)
	return smf.Message([]byte{0xD0 + ch, p})
}

func Test_read_interpretAfterTouchMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"ch0p1": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeAfterTouchMessage(0, 1)},
			WantedRecording: output.WantedRecording{Console: "AfterTouch channel 0 pressure 1\n"},
		},
		"ch1p2": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeAfterTouchMessage(1, 2)},
			WantedRecording: output.WantedRecording{Console: "AfterTouch channel 1 pressure 2\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretAfterTouchMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretAfterTouchMsg() %s", issue)
				}
			}
		})
	}
}

func makeControlChangeMsg(channel, controller, value int) smf.Message {
	ch := byte(channel & 0x0F)
	ctrl := byte(controller & 0x7F)
	val := byte(value & 0x7F)
	return smf.Message([]byte{0xB0 + ch, ctrl, val})
}

func Test_read_interpretControlChangeMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"ch0ctrl1v2": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeControlChangeMsg(0, 1, 2)},
			WantedRecording: output.WantedRecording{Console: "ControlChange channel 0 controller 1 value 2\n"},
		},
		"ch1ctrl2v3": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeControlChangeMsg(1, 2, 3)},
			WantedRecording: output.WantedRecording{Console: "ControlChange channel 1 controller 2 value 3\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretControlChangeMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretControlChangeMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaChannelMsg(channel int) smf.Message {
	return smf.Message([]byte{0xFF, 0x20, 0x01, byte(channel & 0x0F)})
}

func Test_read_interpretMetaChannelMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"ch4": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaChannelMsg(4)},
			WantedRecording: output.WantedRecording{Console: "MetaChannel channel 4\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaChannelMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaChangeMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaCopyrightMsg(message string) smf.Message {
	return makeGenericMetaTextMessage(2, message)
}

func makeGenericMetaTextMessage(messageType int, message string) smf.Message {
	payload := []byte(message)
	msg := []byte{0x0FF, byte(messageType & 0xFF)}
	msg = append(msg, VlqEncode(uint32(len(payload)))...)
	msg = append(msg, payload...)
	return smf.Message(msg)
}

func Test_read_interpretMetaCopyrightMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaCopyrightMsg("(c) foo inc. 1934")},
			WantedRecording: output.WantedRecording{Console: "MetaCopyright text \"(c) foo inc. 1934\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaCopyrightMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaCopyrightMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaCuepointMsg(message string) smf.Message {
	return makeGenericMetaTextMessage(7, message)
}

func Test_read_interpretMetaCuepointMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaCuepointMsg("Soloist starts here")},
			WantedRecording: output.WantedRecording{Console: "MetaCuepoint text \"Soloist starts here\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaCuepointMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaCuepointMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaDeviceMsg(message string) smf.Message {
	return makeGenericMetaTextMessage(9, message)
}

func Test_read_interpretMetaDeviceMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaDeviceMsg("MIDI Port 5")},
			WantedRecording: output.WantedRecording{Console: "MetaDevice text \"MIDI Port 5\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaDeviceMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaDeviceMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaInstrumentMsg(message string) smf.Message {
	return makeGenericMetaTextMessage(4, message)
}

func Test_read_interpretMetaInstrumentMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaInstrumentMsg("Contrabass piccolo")},
			WantedRecording: output.WantedRecording{Console: "MetaInstrument text \"Contrabass piccolo\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaInstrumentMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaInstrumentMsg() %s", issue)
				}
			}
		})
	}
}

func Test_read_interpretSMFFile(t *testing.T) {
	reader := bytes.NewReader(makeTrivialContent())
	data, _ := smf.ReadFrom(reader)
	type args struct {
		data *smf.SMF
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:    &read{key: &smf.Key{IsMajor: true}},
			args: args{data: data},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Quarter note: 120 ticks\n" +
					"16 tracks\n" +
					"Track 0 is empty\n" +
					"Track 1 is empty\n" +
					"Track 2 is empty\n" +
					"Track 3 is empty\n" +
					"Track 4 is empty\n" +
					"Track 5 is empty\n" +
					"Track 6 is empty\n" +
					"Track 7 is empty\n" +
					"Track 8 is empty\n" +
					"Track 9 is empty\n" +
					"Track 10 is empty\n" +
					"Track 11 is empty\n" +
					"Track 12 is empty\n" +
					"Track 13 is empty\n" +
					"Track 14 is empty\n" +
					"Track 15 is empty\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretSMFFile(o, tt.args.data)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretSMFFile() %s", issue)
				}
			}
		})
	}
}

func Test_read_interpretSMFTimeFormat(t *testing.T) {
	type args struct {
		tf smf.TimeFormat
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"expected": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{tf: smf.MetricTicks(120)},
			WantedRecording: output.WantedRecording{Console: "Quarter note: 120 ticks\n"},
		},
		"unusual": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{tf: smf.TimeCode{FramesPerSecond: 29, SubFrames: 40}},
			WantedRecording: output.WantedRecording{Console: "Time: SMPTE30DropFrame 40 subframes\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretSMFTimeFormat(o, tt.args.tf)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretSMFTimeFormat() %s", issue)
				}
			}
		})
	}
}

func makeMetaKeySigMsg(sharps int, major bool) smf.Message {
	content := []byte{0xFF, 0x59, 2, byte(sharps)}
	if major {
		content = append(content, 0)
	} else {
		content = append(content, 1)
	}
	return smf.Message(content)
}

func Test_read_interpretMetaKeySigMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		wantKey *smf.Key
		output.WantedRecording
	}{
		"A-flat minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(-7, false)},
			wantKey:         &smf.Key{Key: 8, Num: 7, IsMajor: false, IsFlat: true},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig A♭Minor (7 flats)\n"},
		},
		"E-flat minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(-6, false)},
			wantKey:         &smf.Key{Key: 3, Num: 6, IsMajor: false, IsFlat: true},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig E♭Minor (6 flats)\n"},
		},
		"B-flat minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(-5, false)},
			wantKey:         &smf.Key{Key: 10, Num: 5, IsMajor: false, IsFlat: true},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig B♭Minor (5 flats)\n"},
		},
		"F minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(-4, false)},
			wantKey:         &smf.Key{Key: 5, Num: 4, IsMajor: false, IsFlat: true},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig FMinor (4 flats)\n"},
		},
		"C minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(-3, false)},
			wantKey:         &smf.Key{Key: 0, Num: 3, IsMajor: false, IsFlat: true},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig CMinor (3 flats)\n"},
		},
		"G minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(-2, false)},
			wantKey:         &smf.Key{Key: 7, Num: 2, IsMajor: false, IsFlat: true},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig GMinor (2 flats)\n"},
		},
		"D minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(-1, false)},
			wantKey:         &smf.Key{Key: 2, Num: 1, IsMajor: false, IsFlat: true},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig DMinor (1 flat)\n"},
		},
		"A minor": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaKeySigMsg(0, false)},
			wantKey:         &smf.Key{Key: 9, Num: 0, IsMajor: false, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig AMinor (0 flats)\n"},
		},
		"C major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(0, true)},
			wantKey:         &smf.Key{Key: 0, Num: 0, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig CMajor (0 sharps)\n"},
		},
		"G major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(1, true)},
			wantKey:         &smf.Key{Key: 7, Num: 1, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig GMajor (1 sharp)\n"},
		},
		"D major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(2, true)},
			wantKey:         &smf.Key{Key: 2, Num: 2, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig DMajor (2 sharps)\n"},
		},
		"A major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(3, true)},
			wantKey:         &smf.Key{Key: 9, Num: 3, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig AMajor (3 sharps)\n"},
		},
		"E major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(4, true)},
			wantKey:         &smf.Key{Key: 4, Num: 4, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig EMajor (4 sharps)\n"},
		},
		"B major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(5, true)},
			wantKey:         &smf.Key{Key: 11, Num: 5, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig BMajor (5 sharps)\n"},
		},
		"F-sharp major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(6, true)},
			wantKey:         &smf.Key{Key: 6, Num: 6, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig F♯Major (6 sharps)\n"},
		},
		"C-sharp major": {
			r:               &read{key: &smf.Key{}},
			args:            args{message: makeMetaKeySigMsg(7, true)},
			wantKey:         &smf.Key{Key: 1, Num: 7, IsMajor: true, IsFlat: false},
			WantedRecording: output.WantedRecording{Console: "MetaKeySig C♯Major (7 sharps)\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaKeySigMsg(o, tt.args.message)
			if gotKey := tt.r.key; !reflect.DeepEqual(gotKey, tt.wantKey) {
				t.Errorf("read.interpretMetaKeySigMsg() got key %#v want key %#v", gotKey, tt.wantKey)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaKeySigMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaLyricMsg(message string) smf.Message {
	return makeGenericMetaTextMessage(5, message)
}

func Test_read_interpretMetaLyricMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaLyricMsg("Tra-la-la")},
			WantedRecording: output.WantedRecording{Console: "MetaLyric text \"Tra-la-la\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaLyricMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaLyricMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaMarkerMessage(message string) smf.Message {
	return makeGenericMetaTextMessage(6, message)
}

func Test_read_interpretMetaMarkerMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaMarkerMessage("Refrain")},
			WantedRecording: output.WantedRecording{Console: "MetaMarker text \"Refrain\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaMarkerMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaMarkerMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaPortMessage(port uint8) smf.Message {
	return smf.Message([]byte{0xff, 33, 1, port})
}

func Test_read_interpretMetaPortMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaPortMessage(42)},
			WantedRecording: output.WantedRecording{Console: "MetaPort port 42\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaPortMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaPortMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaProgramNameMessage(message string) smf.Message {
	return makeGenericMetaTextMessage(8, message)
}

func Test_read_interpretMetaProgramNameMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaProgramNameMessage("Bossa-nova Bassoon")},
			WantedRecording: output.WantedRecording{Console: "MetaProgramName text \"Bossa-nova Bassoon\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaProgramNameMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaProgramNameMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaSMPTEOffsetMessage(frameRate, hour, minute, second, frame, subframe uint8) smf.Message {
	var frameForm uint8
	var maxFrame uint8
	switch frameRate {
	case 1:
		frameForm = 0x20
		maxFrame = 25
	case 2:
		frameForm = 0x40
		maxFrame = 29 // drop 30
	case 3:
		frameForm = 0x60
		maxFrame = 30
	default: // treat as 0
		frameForm = 0
		maxFrame = 24
	}
	if hour > 23 {
		hour = 23
	}
	if minute > 59 {
		minute = 59
	}
	if second > 59 {
		second = 59
	}
	if frame > maxFrame {
		frame = maxFrame
	}
	if subframe > 99 {
		subframe = 99
	}
	return smf.Message([]byte{0xFF, 0x54, 5, frameForm | hour, minute, second, frame, subframe})
}

func Test_read_interpretMetaSMPTEOffsetMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaSMPTEOffsetMessage(0, 1, 2, 3, 4, 5)},
			WantedRecording: output.WantedRecording{Console: "MetaSMPTEOffset 01:02:03 frame 04.05\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaSMPTEOffsetMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaSMPTEOffsetMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaSeqDataMessage(message []byte) smf.Message {
	content := []byte{0xFF, 0x7F}
	content = append(content, VlqEncode(uint32(len(message)))...)
	content = append(content, message...)
	return smf.Message(content)
}

func Test_read_interpretMetaSeqDataMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaSeqDataMessage([]byte{1, 2, 3, 5, 8})},
			WantedRecording: output.WantedRecording{Console: "MetaSeqData bytes [1 2 3 5 8]\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaSeqDataMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaSeqDataMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaSeqNumberMessage(val uint16) smf.Message {
	content := []byte{0xFF, 0, 2}
	content = append(content, encode16(val)...)
	return smf.Message(content)
}

func Test_read_interpretMetaSeqNumberMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaSeqNumberMessage(1234)},
			WantedRecording: output.WantedRecording{Console: "MetaSeqNumber sequence number 1234\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaSeqNumberMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaSeqNumberMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaTempoMessage(value uint32) smf.Message {
	encoded := encode32(value)
	content := []byte{0xFF, 0x51, 3}
	content = append(content, encoded[1:]...)
	return smf.Message(content)
}

func Test_read_interpretMetaTempoMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaTempoMessage(500000)},
			WantedRecording: output.WantedRecording{Console: "MetaTempo bpm 120.000000\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaTempoMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaTempoMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaTextMessage(message string) smf.Message {
	return makeGenericMetaTextMessage(1, message)
}

func Test_read_interpretMetaTextMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaTextMessage("Measure 5")},
			WantedRecording: output.WantedRecording{Console: "MetaText text \"Measure 5\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaTextMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaTextMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaTimeSignatureMessage(numerator, denominatorExponent, tickLength, thirtysecondNotesPerBeat uint8) smf.Message {
	return smf.Message([]byte{0xFF, 0x58, 4, numerator, denominatorExponent, tickLength, thirtysecondNotesPerBeat})
}

func Test_read_interpretMetaTimeSigMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaTimeSignatureMessage(3, 2, 24, 8)},
			WantedRecording: output.WantedRecording{Console: "MetaTimeSig numerator 3 denominator 4 clocksPerClick 24 demiSemiQuaverPerQuarter 8\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaTimeSigMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaTimeSigMsg() %s", issue)
				}
			}
		})
	}
}

func makeMetaTrackNameMessage(message string) smf.Message {
	return makeGenericMetaTextMessage(3, message)
}

func Test_read_interpretMetaTrackNameMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeMetaTrackNameMessage("Lead vocals")},
			WantedRecording: output.WantedRecording{Console: "MetaTrackName text \"Lead vocals\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretMetaTrackNameMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretMetaTrackNameMsg() %s", issue)
				}
			}
		})
	}
}

func makeNoteOffMessage(channel, note, velocity uint8) smf.Message {
	return smf.Message([]byte{0x80 + (channel & 0x0F), note & 0x7F, velocity & 0x7F})
}

func Test_read_interpretNoteOffMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeNoteOffMessage(2, 64, 16)},
			WantedRecording: output.WantedRecording{Console: "NoteOff channel 2 note \"E5\" volume pianississimo (𝆏𝆏𝆏)\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretNoteOffMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretNoteOffMsg() %s", issue)
				}
			}
		})
	}
}

func makeNoteOnMessage(channel, note, velocity uint8) smf.Message {
	return smf.Message([]byte{0x90 + (channel & 0x0F), note & 0x7F, velocity & 0x7F})
}

func Test_read_interpretNoteOnMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeNoteOnMessage(3, 72, 80)},
			WantedRecording: output.WantedRecording{Console: "NoteOn channel 3 note \"C6\" volume mezzo-forte (𝆐𝆑)\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretNoteOnMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretNoteOnMsg() %s", issue)
				}
			}
		})
	}
}

func makePitchBendMessage(channel uint8, value uint16) smf.Message {
	return smf.Message([]byte{0xE0 + (channel & 0x0F), byte(value & 0x7F), byte((value / 128) & 0x7F)})
}

func Test_read_interpretPitchBendMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic1": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makePitchBendMessage(3, 0x00FF)},
			WantedRecording: output.WantedRecording{Console: "PitchBend channel 3 relative -7937 absolute 255\n"},
		},
		"basic2": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makePitchBendMessage(3, 0x07FF)},
			WantedRecording: output.WantedRecording{Console: "PitchBend channel 3 relative -6145 absolute 2047\n"},
		},
		"basic3": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makePitchBendMessage(3, 0x0FFF)},
			WantedRecording: output.WantedRecording{Console: "PitchBend channel 3 relative -4097 absolute 4095\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretPitchBendMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretPitchBendMsg() %s", issue)
				}
			}
		})
	}
}

func makePolyphonicAfterTouchMessage(channel, note, pressure uint8) smf.Message {
	return smf.Message([]byte{0xA0 + (channel % 0x0F), note & 0x7F, pressure & 0x7F})
}

func Test_read_interpretPolyAfterTouchMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makePolyphonicAfterTouchMessage(3, 46, 100)},
			WantedRecording: output.WantedRecording{Console: "PolyAfterTouch channel 3 note A♯3 pressure 100\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretPolyAfterTouchMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretPolyAfterTouchMsg() %s", issue)
				}
			}
		})
	}
}

func makeProgramChangeMessage(channel, instrument uint8) smf.Message {
	return smf.Message([]byte{0xC0 + (channel & 0x0F), instrument % 0x7F})
}

func Test_read_interpretProgramChangeMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"melodic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeProgramChangeMessage(3, 65)},
			WantedRecording: output.WantedRecording{Console: "ProgramChange channel 3 instrument \"Alto sax\"\n"},
		},
		"percussion": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeProgramChangeMessage(9, 65)},
			WantedRecording: output.WantedRecording{Console: "ProgramChange channel 9 instrument \"Low timbale\"\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretProgramChangeMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretProgramChangeMsg() %s", issue)
				}
			}
		})
	}
}

func makeSysExMessage(data []byte) smf.Message {
	content := []byte{0xF0}
	content = append(content, data...)
	content = append(content, 0xF7)
	return smf.Message(content)
}

func makeBusyTrack() smf.Track {
	t := smf.Track{}
	messages := []smf.Message{
		makeAfterTouchMessage(0, 1),
		makeControlChangeMsg(0, 2, 3),
		makeMetaChannelMsg(0),
		makeMetaCopyrightMsg("(c) me 2023"),
		makeMetaCuepointMsg("Soloist start"),
		makeMetaDeviceMsg("device stop"),
		makeMetaInstrumentMsg("bass fife"),
		makeMetaKeySigMsg(1, true),
		makeMetaLyricMsg("foo"),
		makeMetaMarkerMessage("Measure 435"),
		makeMetaPortMessage(3),
		makeMetaProgramNameMessage("warp factor 9"),
		makeMetaSMPTEOffsetMessage(29, 1, 2, 3, 4, 5),
		makeMetaSeqDataMessage([]byte{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}),
		makeMetaSeqNumberMessage(100),
		makeMetaTempoMessage(40000),
		makeMetaTextMessage("hi!"),
		makeMetaTimeSignatureMessage(3, 3, 24, 8),
		makeMetaTrackNameMessage("track 45"),
		makeNoteOffMessage(0, 40, 64),
		makeNoteOnMessage(0, 40, 127),
		makePitchBendMessage(0, 10000),
		makePolyphonicAfterTouchMessage(0, 60, 77),
		makeProgramChangeMessage(0, 33),
		makeSysExMessage([]byte{1, 2, 3, 4, 5}),
		metaEndOfTrackMsg,
	}
	for k, message := range messages {
		t = append(t, smf.Event{Delta: uint32(k), Message: message})
	}
	return t
}

func Test_read_interpretSMFTrack(t *testing.T) {
	type args struct {
		track smf.Track
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"empty": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{track: smf.Track{}},
			WantedRecording: output.WantedRecording{Console: ""},
		},
		"busy": {
			r:    &read{key: &smf.Key{IsMajor: true}},
			args: args{track: makeBusyTrack()},
			WantedRecording: output.WantedRecording{Console: "" +
				"0: delta 0 AfterTouch channel 0 pressure 1\n" +
				"1: delta 1 ControlChange channel 0 controller 2 value 3\n" +
				"2: delta 2 MetaChannel channel 0\n" +
				"3: delta 3 MetaCopyright text \"(c) me 2023\"\n" +
				"4: delta 4 MetaCuepoint text \"Soloist start\"\n" +
				"5: delta 5 MetaDevice text \"device stop\"\n" +
				"6: delta 6 MetaInstrument text \"bass fife\"\n" +
				"7: delta 7 MetaKeySig GMajor (1 sharp)\n" +
				"8: delta 8 MetaLyric text \"foo\"\n" +
				"9: delta 9 MetaMarker text \"Measure 435\"\n" +
				"10: delta 10 MetaPort port 3\n" +
				"11: delta 11 MetaProgramName text \"warp factor 9\"\n" +
				"12: delta 12 MetaSMPTEOffset 01:02:03 frame 04.05\n" +
				"13: delta 13 MetaSeqData bytes [1 1 2 3 5 8 13 21 34 55]\n" +
				"14: delta 14 MetaSeqNumber sequence number 100\n" +
				"15: delta 15 MetaTempo bpm 1500.000000\n" +
				"16: delta 16 MetaText text \"hi!\"\n" +
				"17: delta 17 MetaTimeSig numerator 3 denominator 8 clocksPerClick 24 demiSemiQuaverPerQuarter 8\n" +
				"18: delta 18 MetaTrackName text \"track 45\"\n" +
				"19: delta 19 NoteOff channel 0 note \"E3\" volume mezzo-piano (𝆐𝆏)\n" +
				"20: delta 20 NoteOn channel 0 note \"E3\" volume fortississimo (𝆑𝆑𝆑)\n" +
				"21: delta 21 PitchBend channel 0 relative 1808 absolute 10000\n" +
				"22: delta 22 PolyAfterTouch channel 0 note C5 pressure 77\n" +
				"23: delta 23 ProgramChange channel 0 instrument \"Fingered electric bass\"\n" +
				"24: delta 24 SysEx bytes [1 2 3 4 5]\n" +
				"25: delta 25 Unrecognized message: \"MetaEndOfTrack\" [255 47 0]\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretSMFTrack(o, tt.args.track)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretSMFTrack() %s", issue)
				}
			}
		})
	}
}

func Test_read_interpretSMFTracks(t *testing.T) {
	type args struct {
		tracks []smf.Track
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:    &read{key: &smf.Key{IsMajor: true}},
			args: args{tracks: []smf.Track{{}, makeBusyTrack()}},
			WantedRecording: output.WantedRecording{Console: "" +
				"2 tracks\n" +
				"Track 0 is empty\n" +
				"Track 1:\n" +
				"0: delta 0 AfterTouch channel 0 pressure 1\n" +
				"1: delta 1 ControlChange channel 0 controller 2 value 3\n" +
				"2: delta 2 MetaChannel channel 0\n" +
				"3: delta 3 MetaCopyright text \"(c) me 2023\"\n" +
				"4: delta 4 MetaCuepoint text \"Soloist start\"\n" +
				"5: delta 5 MetaDevice text \"device stop\"\n" +
				"6: delta 6 MetaInstrument text \"bass fife\"\n" +
				"7: delta 7 MetaKeySig GMajor (1 sharp)\n" +
				"8: delta 8 MetaLyric text \"foo\"\n" +
				"9: delta 9 MetaMarker text \"Measure 435\"\n" +
				"10: delta 10 MetaPort port 3\n" +
				"11: delta 11 MetaProgramName text \"warp factor 9\"\n" +
				"12: delta 12 MetaSMPTEOffset 01:02:03 frame 04.05\n" +
				"13: delta 13 MetaSeqData bytes [1 1 2 3 5 8 13 21 34 55]\n" +
				"14: delta 14 MetaSeqNumber sequence number 100\n" +
				"15: delta 15 MetaTempo bpm 1500.000000\n" +
				"16: delta 16 MetaText text \"hi!\"\n" +
				"17: delta 17 MetaTimeSig numerator 3 denominator 8 clocksPerClick 24 demiSemiQuaverPerQuarter 8\n" +
				"18: delta 18 MetaTrackName text \"track 45\"\n" +
				"19: delta 19 NoteOff channel 0 note \"E3\" volume mezzo-piano (𝆐𝆏)\n" +
				"20: delta 20 NoteOn channel 0 note \"E3\" volume fortississimo (𝆑𝆑𝆑)\n" +
				"21: delta 21 PitchBend channel 0 relative 1808 absolute 10000\n" +
				"22: delta 22 PolyAfterTouch channel 0 note C5 pressure 77\n" +
				"23: delta 23 ProgramChange channel 0 instrument \"Fingered electric bass\"\n" +
				"24: delta 24 SysEx bytes [1 2 3 4 5]\n" +
				"25: delta 25 Unrecognized message: \"MetaEndOfTrack\" [255 47 0]\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretSMFTracks(o, tt.args.tracks)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretSMFTracks() %s", issue)
				}
			}
		})
	}
}

func Test_read_interpretSysExMsg(t *testing.T) {
	type args struct {
		message smf.Message
	}
	tests := map[string]struct {
		r *read
		args
		output.WantedRecording
	}{
		"basic": {
			r:               &read{key: &smf.Key{IsMajor: true}},
			args:            args{message: makeSysExMessage([]byte{1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233})},
			WantedRecording: output.WantedRecording{Console: "SysEx bytes [1 1 2 3 5 8 13 21 34 55 89 144 233]\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.r.interpretSysExMsg(o, tt.args.message)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("read.interpretSysExMsg() %s", issue)
				}
			}
		})
	}
}

func Test_read_asInstrument(t *testing.T) {
	type args struct {
		channel uint8
		program uint8
	}
	tests := map[string]struct {
		r *read
		args
		want string
	}{
		"Acoustic grand piano":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x00}, want: "Acoustic grand piano"},
		"Bright acoustic piano":            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x01}, want: "Bright acoustic piano"},
		"Electric grand piano":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x02}, want: "Electric grand piano"},
		"Honky tonk piano":                 {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x03}, want: "Honky tonk piano"},
		"Electric piano 1":                 {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x04}, want: "Electric piano 1"},
		"Electric piano 2":                 {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x05}, want: "Electric piano 2"},
		"Harpsichord":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x06}, want: "Harpsicord"},
		"Clavinet":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x07}, want: "Clavinet"},
		"Celesta":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x08}, want: "Celesta"},
		"Glockenspiel":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x09}, want: "Glockenspiel"},
		"Music box":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x0A}, want: "Music box"},
		"Vibraphone":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x0B}, want: "Vibraphone"},
		"Marimba":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x0C}, want: "Marimba"},
		"Xylophone":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x0D}, want: "Xylophone"},
		"Tubular bell":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x0E}, want: "Tubular bell"},
		"Dulcimer":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x0F}, want: "Dulcimer"},
		"Hammond / drawbar organ":          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x10}, want: "Hammond / drawbar organ"},
		"Percussive organ":                 {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x11}, want: "Percussive organ"},
		"Rock organ":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x12}, want: "Rock organ"},
		"Church organ":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x13}, want: "Church organ"},
		"Reed organ":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x14}, want: "Reed organ"},
		"Accordion":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x15}, want: "Accordion"},
		"Harmonica":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x16}, want: "Harmonica"},
		"Tango accordion":                  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x17}, want: "Tango accordion"},
		"Nylon string acoustic guitar":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x18}, want: "Nylon string acoustic guitar"},
		"Steel string acoustic guitar":     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x19}, want: "Steel string acoustic guitar"},
		"Jazz electric guitar":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x1A}, want: "Jazz electric guitar"},
		"Clean electric guitar":            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x1B}, want: "Clean electric guitar"},
		"Muted electric guitar":            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x1C}, want: "Muted electric guitar"},
		"Overdriven guitar":                {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x1D}, want: "Overdriven guitar"},
		"Distortion guitar":                {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x1E}, want: "Distortion guitar"},
		"Guitar harmonics":                 {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x1F}, want: "Guitar harmonics"},
		"Acoustic bass":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x20}, want: "Acoustic bass"},
		"Fingered electric bass":           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x21}, want: "Fingered electric bass"},
		"Picked electric bass":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x22}, want: "Picked electric bass"},
		"Fretless bass":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x23}, want: "Fretless bass"},
		"Slap bass 1":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x24}, want: "Slap bass 1"},
		"Slap bass 2":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x25}, want: "Slap bass 2"},
		"Synth bass 1":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x26}, want: "Synth bass 1"},
		"Synth bass 2":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x27}, want: "Synth bass 2"},
		"Violin":                           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x28}, want: "Violin"},
		"Viola":                            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x29}, want: "Viola"},
		"Cello":                            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x2A}, want: "Cello"},
		"Contrabass":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x2B}, want: "Contrabass"},
		"Tremolo strings":                  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x2C}, want: "Tremolo strings"},
		"Pizzicato strings":                {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x2D}, want: "Pizzicato strings"},
		"Orchestral strings / harp":        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x2E}, want: "Orchestral strings / harp"},
		"Timpani":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x2F}, want: "Timpani"},
		"String ensemble 1":                {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x30}, want: "String ensemble 1"},
		"String ensemble 2 / slow strings": {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x31}, want: "String ensemble 2 / slow strings"},
		"Synth strings 1":                  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x32}, want: "Synth strings 1"},
		"Synth strings 2":                  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x33}, want: "Synth strings 2"},
		"Choir aahs":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x34}, want: "Choir aahs"},
		"Voice oohs":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x35}, want: "Voice oohs"},
		"Synth choir / voice":              {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x36}, want: "Synth choir / voice"},
		"Orchestra hit":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x37}, want: "Orchestra hit"},
		"Trumpet":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x38}, want: "Trumpet"},
		"Trombone":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x39}, want: "Trombone"},
		"Tuba":                             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x3A}, want: "Tuba"},
		"Muted trumpet":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x3B}, want: "Muted trumpet"},
		"French horn":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x3C}, want: "French horn"},
		"Brass ensemble":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x3D}, want: "Brass ensemble"},
		"Synth brass 1":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x3E}, want: "Synth brass 1"},
		"Synth brass 2":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x3F}, want: "Synth brass 2"},
		"Soprano sax":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x40}, want: "Soprano sax"},
		"Alto sax":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x41}, want: "Alto sax"},
		"Tenor sax":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x42}, want: "Tenor sax"},
		"Baritone sax":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x43}, want: "Baritone sax"},
		"Oboe":                             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x44}, want: "Oboe"},
		"English horn":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x45}, want: "English horn"},
		"Bassoon":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x46}, want: "Bassoon"},
		"Clarinet":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x47}, want: "Clarinet"},
		"Piccolo":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x48}, want: "Piccolo"},
		"Flute":                            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x49}, want: "Flute"},
		"Recorder":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x4A}, want: "Recorder"},
		"Pan flute":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x4B}, want: "Pan flute"},
		"Bottle blow / blown bottle":       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x4C}, want: "Bottle blow / blown bottle"},
		"Shakuhachi":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x4D}, want: "Shakuhachi"},
		"Whistle":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x4E}, want: "Whistle"},
		"Ocarina":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x4F}, want: "Ocarina"},
		"Synth square wave":                {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x50}, want: "Synth square wave"},
		"Synth saw wave":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x51}, want: "Synth saw wave"},
		"Synth calliope":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x52}, want: "Synth calliope"},
		"Synth chiff":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x53}, want: "Synth chiff"},
		"Synth charang":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x54}, want: "Synth charang"},
		"Synth voice":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x55}, want: "Synth voice"},
		"Synth fifths saw":                 {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x56}, want: "Synth fifths saw"},
		"Synth brass and lead":             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x57}, want: "Synth brass and lead"},
		"Fantasia / new age":               {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x58}, want: "Fantasia / new age"},
		"Warm pad":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x59}, want: "Warm pad"},
		"Polysynth":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x5A}, want: "Polysynth"},
		"Space vox / choir":                {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x5B}, want: "Space vox / choir"},
		"Bowed glass":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x5C}, want: "Bowed glass"},
		"Metal pad":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x5D}, want: "Metal pad"},
		"Halo pad":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x5E}, want: "Halo pad"},
		"Sweep pad":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x5F}, want: "Sweep pad"},
		"Ice rain":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x60}, want: "Ice rain"},
		"Soundtrack":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x61}, want: "Soundtrack"},
		"Crystal":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x62}, want: "Crystal"},
		"Atmosphere":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x63}, want: "Atmosphere"},
		"Brightness":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x64}, want: "Brightness"},
		"Goblins":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x65}, want: "Goblins"},
		"Echo drops / echoes":              {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x66}, want: "Echo drops / echoes"},
		"Sci fi":                           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x67}, want: "Sci fi"},
		"Sitar":                            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x68}, want: "Sitar"},
		"Banjo":                            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x69}, want: "Banjo"},
		"Shamisen":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x6A}, want: "Shamisen"},
		"Koto":                             {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x6B}, want: "Koto"},
		"Kalimba":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x6C}, want: "Kalimba"},
		"Bag pipe":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x6D}, want: "Bag pipe"},
		"Fiddle":                           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x6E}, want: "Fiddle"},
		"Shanai":                           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x6F}, want: "Shanai"},
		"Tinkle bell":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x70}, want: "Tinkle bell"},
		"Agogo":                            {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x71}, want: "Agogo"},
		"Steel drums":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x72}, want: "Steel drums"},
		"Woodblock":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x73}, want: "Woodblock"},
		"Taiko drum":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x74}, want: "Taiko drum"},
		"Melodic tom":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x75}, want: "Melodic tom"},
		"Synth drum":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x76}, want: "Synth drum"},
		"Reverse cymbal":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x77}, want: "Reverse cymbal"},
		"Guitar fret noise":                {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x78}, want: "Guitar fret noise"},
		"Breath noise":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x79}, want: "Breath noise"},
		"Seashore":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x7A}, want: "Seashore"},
		"Bird tweet":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x7B}, want: "Bird tweet"},
		"Telephone ring":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x7C}, want: "Telephone ring"},
		"Helicopter":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x7D}, want: "Helicopter"},
		"Applause":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x7E}, want: "Applause"},
		"Gunshot":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 0, program: 0x7F}, want: "Gunshot"},
		"Unknown percussion 0x21":          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x21}, want: "Unknown percussion instrument 33"},
		"Acoustic bass drum":               {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x22}, want: "Acoustic bass drum"},
		"Bass drum 1":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x23}, want: "Bass drum 1"},
		"Side stick":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x24}, want: "Side stick"},
		"Acoustic snare":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x25}, want: "Acoustic snare"},
		"Hand clap":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x26}, want: "Hand clap"},
		"Electric snare":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x27}, want: "Electric snare"},
		"Low floor tom":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x28}, want: "Low floor tom"},
		"Closed hihat":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x29}, want: "Closed hihat"},
		"High floor tom":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x2A}, want: "High floor tom"},
		"Pedal hihat":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x2B}, want: "Pedal hihat"},
		"Low tom":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x2C}, want: "Low tom"},
		"Open hihat":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x2D}, want: "Open hihat"},
		"Low-mid tom":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x2E}, want: "Low-mid tom"},
		"High-mid tom":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x2F}, want: "High-mid tom"},
		"Crash cymbal 1":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x30}, want: "Crash cymbal 1"},
		"High tom":                         {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x31}, want: "High tom"},
		"Ride cymbal 1":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x32}, want: "Ride cymbal 1"},
		"Chinese cymbal":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x33}, want: "Chinese cymbal"},
		"Ride bell":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x34}, want: "Ride bell"},
		"Tambourine":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x35}, want: "Tambourine"},
		"Splash cymbal":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x36}, want: "Splash cymbal"},
		"Cowbell":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x37}, want: "Cowbell"},
		"Crash cymbal 2":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x38}, want: "Crash cymbal 2"},
		"Vibraslap":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x39}, want: "Vibraslap"},
		"Ride cymbal 2":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x3A}, want: "Ride cymbal 2"},
		"High bongo":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x3B}, want: "High bongo"},
		"Low bongo":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x3C}, want: "Low bongo"},
		"Mute high conga":                  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x3D}, want: "Mute high conga"},
		"Open high conga":                  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x3E}, want: "Open high conga"},
		"Low conga":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x3F}, want: "Low conga"},
		"High timbale":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x40}, want: "High timbale"},
		"Low timbale":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x41}, want: "Low timbale"},
		"High agogo":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x42}, want: "High agogo"},
		"Low agogo":                        {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x43}, want: "Low agogo"},
		"Cabasa":                           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x44}, want: "Cabasa"},
		"Maracas":                          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x45}, want: "Maracas"},
		"Short whistle":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x46}, want: "Short whistle"},
		"Long whistle":                     {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x47}, want: "Long whistle"},
		"Short guiro":                      {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x48}, want: "Short guiro"},
		"Long guiro":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x49}, want: "Long guiro"},
		"Claves":                           {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x4A}, want: "Claves"},
		"High wood block":                  {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x4B}, want: "High wood block"},
		"Low wood block":                   {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x4C}, want: "Low wood block"},
		"Mute cuica":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x4D}, want: "Mute cuica"},
		"Open cuica":                       {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x4E}, want: "Open cuica"},
		"Mute triangle":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x4F}, want: "Mute triangle"},
		"Open triangle":                    {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x50}, want: "Open triangle"},
		"Unknown percussion 0x51":          {r: &read{key: &smf.Key{IsMajor: true}}, args: args{channel: 9, program: 0x51}, want: "Unknown percussion instrument 81"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.r.asInstrument(tt.args.channel, tt.args.program); got != tt.want {
				t.Errorf("read.asInstrument() = %v, want %v", got, tt.want)
			}
		})
	}
}
