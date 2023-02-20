package commands

import (
	"flag"
	"fmt"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func init() {
	tools.AddCommandData(readCommandName, &tools.CommandDescription{IsDefault: IsDefault(readCommandName), Initializer: newRead})
}

const (
	readCommandName = "read"
)

var (
	majorKeys = map[uint8]string{
		0:  "C",
		1:  "C♯",
		2:  "D",
		3:  "D♯",
		4:  "E",
		5:  "F",
		6:  "F♯",
		7:  "G",
		8:  "G♯",
		9:  "A",
		10: "A♯",
		11: "B",
	}
	minorKeys = map[uint8]string{
		0:  "C",
		1:  "D♭",
		2:  "D",
		3:  "E♭",
		4:  "E",
		5:  "F",
		6:  "G♭",
		7:  "G",
		8:  "A♭",
		9:  "A",
		10: "B♭",
		11: "B",
	}
	percussion = map[uint8]string{
		0x22: "Acoustic bass drum",
		0x23: "Bass drum 1",
		0x24: "Side stick",
		0x25: "Acoustic snare",
		0x26: "Hand clap",
		0x27: "Electric snare",
		0x28: "Low floor tom",
		0x29: "Closed hihat",
		0x2A: "High floor tom",
		0x2B: "Pedal hihat",
		0x2C: "Low tom",
		0x2D: "Open hihat",
		0x2E: "Low-mid tom",
		0x2F: "High-mid tom",
		0x30: "Crash cymbal 1",
		0x31: "High tom",
		0x32: "Ride cymbal 1",
		0x33: "Chinese cymbal",
		0x34: "Ride bell",
		0x35: "Tambourine",
		0x36: "Splash cymbal",
		0x37: "Cowbell",
		0x38: "Crash cymbal 2",
		0x39: "Vibraslap",
		0x3A: "Ride cymbal 2",
		0x3B: "High bongo",
		0x3C: "Low bongo",
		0x3D: "Mute high conga",
		0x3E: "Open high conga",
		0x3F: "Low conga",
		0x40: "High timbale",
		0x41: "Low timbale",
		0x42: "High agogo",
		0x43: "Low agogo",
		0x44: "Cabasa",
		0x45: "Maracas",
		0x46: "Short whistle",
		0x47: "Long whistle",
		0x48: "Short guiro",
		0x49: "Long guiro",
		0x4A: "Claves",
		0x4B: "High wood block",
		0x4C: "Low wood block",
		0x4D: "Mute cuica",
		0x4E: "Open cuica",
		0x4F: "Mute triangle",
		0x50: "Open triangle",
	}
	instruments = map[uint8]string{
		0x00: "Acoustic grand piano",
		0x01: "Bright acoustic piano",
		0x02: "Electric grand piano",
		0x03: "Honky tonk piano",
		0x04: "Electric piano 1",
		0x05: "Electric piano 2",
		0x06: "Harpsicord",
		0x07: "Clavinet",
		0x08: "Celesta",
		0x09: "Glockenspiel",
		0x0A: "Music box",
		0x0B: "Vibraphone",
		0x0C: "Marimba",
		0x0D: "Xylophone",
		0x0E: "Tubular bell",
		0x0F: "Dulcimer",
		0x10: "Hammond / drawbar organ",
		0x11: "Percussive organ",
		0x12: "Rock organ",
		0x13: "Church organ",
		0x14: "Reed organ",
		0x15: "Accordion",
		0x16: "Harmonica",
		0x17: "Tango accordion",
		0x18: "Nylon string acoustic guitar",
		0x19: "Steel string acoustic guitar",
		0x1A: "Jazz electric guitar",
		0x1B: "Clean electric guitar",
		0x1C: "Muted electric guitar",
		0x1D: "Overdriven guitar",
		0x1E: "Distortion guitar",
		0x1F: "Guitar harmonics",
		0x20: "Acoustic bass",
		0x21: "Fingered electric bass",
		0x22: "Picked electric bass",
		0x23: "Fretless bass",
		0x24: "Slap bass 1",
		0x25: "Slap bass 2",
		0x26: "Synth bass 1",
		0x27: "Synth bass 2",
		0x28: "Violin",
		0x29: "Viola",
		0x2A: "Cello",
		0x2B: "Contrabass",
		0x2C: "Tremolo strings",
		0x2D: "Pizzicato strings",
		0x2E: "Orchestral strings / harp",
		0x2F: "Timpani",
		0x30: "String ensemble 1",
		0x31: "String ensemble 2 / slow strings",
		0x32: "Synth strings 1",
		0x33: "Synth strings 2",
		0x34: "Choir aahs",
		0x35: "Voice oohs",
		0x36: "Synth choir / voice",
		0x37: "Orchestra hit",
		0x38: "Trumpet",
		0x39: "Trombone",
		0x3A: "Tuba",
		0x3B: "Muted trumpet",
		0x3C: "French horn",
		0x3D: "Brass ensemble",
		0x3E: "Synth brass 1",
		0x3F: "Synth brass 2",
		0x40: "Soprano sax",
		0x41: "Alto sax",
		0x42: "Tenor sax",
		0x43: "Baritone sax",
		0x44: "Oboe",
		0x45: "English horn",
		0x46: "Bassoon",
		0x47: "Clarinet",
		0x48: "Piccolo",
		0x49: "Flute",
		0x4A: "Recorder",
		0x4B: "Pan flute",
		0x4C: "Bottle blow / blown bottle",
		0x4D: "Shakuhachi",
		0x4E: "Whistle",
		0x4F: "Ocarina",
		0x50: "Synth square wave",
		0x51: "Synth saw wave",
		0x52: "Synth calliope",
		0x53: "Synth chiff",
		0x54: "Synth charang",
		0x55: "Synth voice",
		0x56: "Synth fifths saw",
		0x57: "Synth brass and lead",
		0x58: "Fantasia / new age",
		0x59: "Warm pad",
		0x5A: "Polysynth",
		0x5B: "Space vox / choir",
		0x5C: "Bowed glass",
		0x5D: "Metal pad",
		0x5E: "Halo pad",
		0x5F: "Sweep pad",
		0x60: "Ice rain",
		0x61: "Soundtrack",
		0x62: "Crystal",
		0x63: "Atmosphere",
		0x64: "Brightness",
		0x65: "Goblins",
		0x66: "Echo drops / echoes",
		0x67: "Sci fi",
		0x68: "Sitar",
		0x69: "Banjo",
		0x6A: "Shamisen",
		0x6B: "Koto",
		0x6C: "Kalimba",
		0x6D: "Bag pipe",
		0x6E: "Fiddle",
		0x6F: "Shanai",
		0x70: "Tinkle bell",
		0x71: "Agogo",
		0x72: "Steel drums",
		0x73: "Woodblock",
		0x74: "Taiko drum",
		0x75: "Melodic tom",
		0x76: "Synth drum",
		0x77: "Reverse cymbal",
		0x78: "Guitar fret noise",
		0x79: "Breath noise",
		0x7A: "Seashore",
		0x7B: "Bird tweet",
		0x7C: "Telephone ring",
		0x7D: "Helicopter",
		0x7E: "Applause",
		0x7F: "Gunshot",
	}
)

type read struct {
	key *smf.Key
}

func newRead(o output.Bus, c *tools.Configuration, flags *flag.FlagSet) (tools.CommandProcessor, bool) {
	return newReadCommand(o, c, flags)
}

func newReadCommand(o output.Bus, c *tools.Configuration, flags *flag.FlagSet) (tools.CommandProcessor, bool) {
	return &read{key: &smf.Key{Key: 0, Num: 0, IsMajor: true, IsFlat: false}}, true
}

func (r *read) Exec(o output.Bus, args []string) (ok bool) {
	if len(args) == 0 {
		tools.ReportNothingToDo(o, readCommandName, nil)
	} else {
		ok = true // optimistic!
		for _, arg := range args {
			if data, err := smf.ReadFile(arg); err != nil {
				o.WriteCanonicalError("An error occurred while reading %q: %v", arg, err)
				ok = false
			} else {
				o.WriteConsole("File: %q\n", arg)
				r.interpretSMFFile(o, data)
				o.WriteConsole("EOF %q\n", arg)
			}
		}
	}
	return
}

func (r *read) asNote(raw uint8) string {
	noteMap := minorKeys
	if r.key.IsMajor {
		noteMap = majorKeys
	}
	return fmt.Sprintf("%s%d", noteMap[raw%12], raw/12)
}

func (r *read) asVolume(velocity uint8) string {
	switch velocity {
	case 16:
		return "pianississimo (𝆏𝆏𝆏)"
	case 33:
		return "pianissimo (𝆏𝆏)"
	case 49:
		return "piano (𝆏)"
	case 64:
		return "mezzo-piano (𝆐𝆏)"
	case 80:
		return "mezzo-forte (𝆐𝆑)"
	case 96:
		return "forte (𝆑)"
	case 112:
		return "fortissimo (𝆑𝆑)"
	case 127:
		return "fortississimo (𝆑𝆑𝆑)"
	default:
		switch {
		case velocity < 16:
			return fmt.Sprintf("below pianississimo (𝆏𝆏𝆏) (%d)", velocity)
		case velocity < 33:
			return fmt.Sprintf("between pianississimo (𝆏𝆏𝆏) and pianissimo (𝆏𝆏) (%d)", velocity)
		case velocity < 49:
			return fmt.Sprintf("between pianissimo (𝆏𝆏) and piano (𝆏) (%d)", velocity)
		case velocity < 64:
			return fmt.Sprintf("between piano (𝆏) and mezzo-piano (𝆐𝆏) (%d)", velocity)
		case velocity < 80:
			return fmt.Sprintf("between mezzo-piano (𝆐𝆏) and mezzo-forte (𝆐𝆑) (%d)", velocity)
		case velocity < 96:
			return fmt.Sprintf("between mezzo-forte (𝆐𝆑) and forte (𝆑) (%d)", velocity)
		case velocity < 112:
			return fmt.Sprintf("between forte (𝆑) and fortissimo (𝆑𝆑) (%d)", velocity)
		case velocity < 127:
			return fmt.Sprintf("between fortissimo (𝆑𝆑) and fortississimo (𝆑𝆑𝆑) (%d)", velocity)
		default:
			return fmt.Sprintf("above fortississimo (𝆑𝆑𝆑) (%d)", velocity)
		}
	}
}

func (r *read) interpretAfterTouchMsg(o output.Bus, message smf.Message) {
	var channel, pressure uint8
	_ = message.GetAfterTouch(&channel, &pressure)
	o.WriteConsole("AfterTouch channel %d pressure %d\n", channel, pressure)
}

func (r *read) interpretControlChangeMsg(o output.Bus, message smf.Message) {
	var channel, controller, value uint8
	_ = message.GetControlChange(&channel, &controller, &value)
	o.WriteConsole("ControlChange channel %d controller %d value %d\n", channel, controller, value)
}

func (r *read) interpretMetaChannelMsg(o output.Bus, message smf.Message) {
	var channel uint8
	_ = message.GetMetaChannel(&channel)
	o.WriteConsole("MetaChannel channel %d\n", channel)
}

func (r *read) interpretMetaCopyrightMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaCopyright(&text)
	o.WriteConsole("MetaCopyright text %q\n", text)
}

func (r *read) interpretMetaCuepointMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaCuepoint(&text)
	o.WriteConsole("MetaCuepoint text %q\n", text)
}

func (r *read) interpretMetaDeviceMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaDevice(&text)
	o.WriteConsole("MetaDevice text %q\n", text)
}

func (r *read) interpretMetaInstrumentMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaInstrument(&text)
	o.WriteConsole("MetaInstrument text %q\n", text)
}

func (r *read) interpretSMFFile(o output.Bus, data *smf.SMF) {
	r.interpretSMFTimeFormat(o, data.TimeFormat)
	r.interpretSMFTracks(o, data.Tracks)
}

func (r *read) interpretSMFTimeFormat(o output.Bus, tf smf.TimeFormat) {
	if mt, ok := tf.(smf.MetricTicks); ok {
		o.WriteConsole("Quarter note: %d ticks\n", mt.Ticks4th())
	} else {
		o.WriteConsole("Time: %s\n", tf)
	}
}

func (r *read) interpretMetaKeySigMsg(o output.Bus, message smf.Message) {
	_ = message.GetMetaKey(r.key)
	delta := "flats"
	if r.key.Num == 1 {
		delta = "flat"
	}
	noteMap := minorKeys
	modifier := "Minor"
	if r.key.IsMajor {
		delta = "sharps"
		if r.key.Num == 1 {
			delta = "sharp"
		}
		noteMap = majorKeys
		modifier = "Major"
	}
	o.WriteConsole("MetaKeySig %s%s (%d %s)\n", noteMap[r.key.Key], modifier, r.key.Num, delta)
}

func (r *read) interpretMetaLyricMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaLyric(&text)
	o.WriteConsole("MetaLyric text %q\n", text)
}

func (r *read) interpretMetaMarkerMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaMarker(&text)
	o.WriteConsole("MetaMarker text %q\n", text)
}

func (r *read) interpretMetaPortMsg(o output.Bus, message smf.Message) {
	var port uint8
	_ = message.GetMetaPort(&port)
	o.WriteConsole("MetaPort port %d\n", port)
}

func (r *read) interpretMetaProgramNameMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaProgramName(&text)
	o.WriteConsole("MetaProgramName text %q\n", text)
}

func (r *read) interpretMetaSMPTEOffsetMsg(o output.Bus, message smf.Message) {
	var hour, minute, second, frame, fractFrame uint8
	_ = message.GetMetaSMPTEOffsetMsg(&hour, &minute, &second, &frame, &fractFrame)
	o.WriteConsole("MetaSMPTEOffset %02d:%02d:%02d frame %02d.%.02d\n", hour, minute, second, frame, fractFrame)
}

func (r *read) interpretMetaSeqDataMsg(o output.Bus, message smf.Message) {
	var bt []byte
	_ = message.GetMetaSeqData(&bt)
	o.WriteConsole("MetaSeqData bytes %v\n", bt)
}

func (r *read) interpretMetaSeqNumberMsg(o output.Bus, message smf.Message) {
	var sequenceNumber uint16
	_ = message.GetMetaSeqNumber(&sequenceNumber)
	o.WriteConsole("MetaSeqNumber sequence number %d\n", sequenceNumber)
}

func (r *read) interpretMetaTempoMsg(o output.Bus, message smf.Message) {
	var bpm float64
	_ = message.GetMetaTempo(&bpm)
	o.WriteConsole("MetaTempo bpm %f\n", bpm)
}

func (r *read) interpretMetaTextMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaText(&text)
	o.WriteConsole("MetaText text %q\n", text)
}

func (r *read) interpretMetaTimeSigMsg(o output.Bus, message smf.Message) {
	var numerator, denominator, clocksPerClick, demiSemiQuaverPerQuarter uint8
	_ = message.GetMetaTimeSig(&numerator, &denominator, &clocksPerClick, &demiSemiQuaverPerQuarter)
	o.WriteConsole("MetaTimeSig numerator %d denominator %d clocksPerClick %d demiSemiQuaverPerQuarter %d\n", numerator, denominator, clocksPerClick, demiSemiQuaverPerQuarter)
}

func (r *read) interpretMetaTrackNameMsg(o output.Bus, message smf.Message) {
	var text string
	_ = message.GetMetaTrackName(&text)
	o.WriteConsole("MetaTrackName text %q\n", text)
}

func (r *read) interpretNoteOffMsg(o output.Bus, message smf.Message) {
	var channel, key, velocity uint8
	_ = message.GetNoteOff(&channel, &key, &velocity)
	o.WriteConsole("NoteOff channel %d note %q volume %s\n", channel, r.asNote(key), r.asVolume(velocity))
}

func (r *read) interpretNoteOnMsg(o output.Bus, message smf.Message) {
	var channel, key, velocity uint8
	_ = message.GetNoteOn(&channel, &key, &velocity)
	o.WriteConsole("NoteOn channel %d note %q volume %s\n", channel, r.asNote(key), r.asVolume(velocity))
}

func (r *read) interpretPitchBendMsg(o output.Bus, message smf.Message) {
	var channel uint8
	var relative int16
	var absolute uint16
	_ = message.GetPitchBend(&channel, &relative, &absolute)
	o.WriteConsole("PitchBend channel %d relative %d absolute %d\n", channel, relative, absolute)
}

func (r *read) interpretPolyAfterTouchMsg(o output.Bus, message smf.Message) {
	var channel, key, pressure uint8
	_ = message.GetPolyAfterTouch(&channel, &key, &pressure)
	o.WriteConsole("PolyAfterTouch channel %d note %s pressure %d\n", channel, r.asNote(key), pressure)
}

func (r *read) asInstrument(channel, program uint8) string {
	switch channel {
	case 9:
		if s, ok := percussion[program&0x7F]; ok {
			return s
		}
		return fmt.Sprintf("Unknown percussion instrument %d", program)
	default:
		return instruments[program&0x7F]
	}
}

func (r *read) interpretProgramChangeMsg(o output.Bus, message smf.Message) {
	var channel, program uint8
	_ = message.GetProgramChange(&channel, &program)
	o.WriteConsole("ProgramChange channel %d instrument %q\n", channel, r.asInstrument(channel, program))
}

func (r *read) interpretSMFTrack(o output.Bus, track smf.Track) {
	for k, event := range track {
		o.WriteConsole("%d: delta %d ", k, event.Delta)
		message := event.Message
		switch message.Type() {
		case midi.AfterTouchMsg:
			r.interpretAfterTouchMsg(o, message)
		case midi.ControlChangeMsg:
			r.interpretControlChangeMsg(o, message)
		case smf.MetaChannelMsg:
			r.interpretMetaChannelMsg(o, message)
		case smf.MetaCopyrightMsg:
			r.interpretMetaCopyrightMsg(o, message)
		case smf.MetaCuepointMsg:
			r.interpretMetaCuepointMsg(o, message)
		case smf.MetaDeviceMsg:
			r.interpretMetaDeviceMsg(o, message)
		case smf.MetaInstrumentMsg:
			r.interpretMetaInstrumentMsg(o, message)
		case smf.MetaKeySigMsg:
			r.interpretMetaKeySigMsg(o, message)
		case smf.MetaLyricMsg:
			r.interpretMetaLyricMsg(o, message)
		case smf.MetaMarkerMsg:
			r.interpretMetaMarkerMsg(o, message)
		case smf.MetaPortMsg:
			r.interpretMetaPortMsg(o, message)
		case smf.MetaProgramNameMsg:
			r.interpretMetaProgramNameMsg(o, message)
		case smf.MetaSMPTEOffsetMsg:
			r.interpretMetaSMPTEOffsetMsg(o, message)
		case smf.MetaSeqDataMsg:
			r.interpretMetaSeqDataMsg(o, message)
		case smf.MetaSeqNumberMsg:
			r.interpretMetaSeqNumberMsg(o, message)
		case smf.MetaTempoMsg:
			r.interpretMetaTempoMsg(o, message)
		case smf.MetaTextMsg:
			r.interpretMetaTextMsg(o, message)
		case smf.MetaTimeSigMsg:
			r.interpretMetaTimeSigMsg(o, message)
		case smf.MetaTrackNameMsg:
			r.interpretMetaTrackNameMsg(o, message)
		case midi.NoteOffMsg:
			r.interpretNoteOffMsg(o, message)
		case midi.NoteOnMsg:
			r.interpretNoteOnMsg(o, message)
		case midi.PitchBendMsg:
			r.interpretPitchBendMsg(o, message)
		case midi.PolyAfterTouchMsg:
			r.interpretPolyAfterTouchMsg(o, message)
		case midi.ProgramChangeMsg:
			r.interpretProgramChangeMsg(o, message)
		case midi.SysExMsg:
			r.interpretSysExMsg(o, message)
		default:
			o.WriteConsole("Unrecognized message: %q %v\n", message.Type(), message.Bytes())
		}
	}
}

func (r *read) interpretSMFTracks(o output.Bus, tracks []smf.Track) {
	o.WriteConsole("%d tracks\n", len(tracks))
	for k, track := range tracks {
		if track.IsEmpty() {
			o.WriteConsole("Track %d is empty\n", k)
		} else {
			o.WriteConsole("Track %d:\n", k)
			r.interpretSMFTrack(o, track)
		}
	}
}

func (r *read) interpretSysExMsg(o output.Bus, message smf.Message) {
	var bt []byte
	_ = message.GetSysEx(&bt)
	o.WriteConsole("SysEx bytes %v\n", bt)
}
