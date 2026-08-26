package naming

import "testing"

func TestRenderDefaultTemplate(t *testing.T) {
	tpl, err := Parse(Default)
	if err != nil {
		t.Fatalf("Parse(Default): %v", err)
	}

	got := tpl.Render(Fields{
		AlbumArtist: "Cat Power",
		Album:       "Jukebox",
		Title:       "Sea of Love",
		Year:        2008,
		Disc:        1,
		Track:       3,
		Ext:         "mp3",
	})

	want := "Cat Power/2008 - Jukebox/01-03 Sea of Love.mp3"
	if got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRenderZeroPadsWiderNumbers(t *testing.T) {
	tpl, err := Parse("{track:03}.{ext}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := tpl.Render(Fields{Track: 7, Ext: "flac"}), "007.flac"; got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRenderUnknownNumericFieldIsBlank(t *testing.T) {
	tpl, err := Parse("[{year}] {title}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := tpl.Render(Fields{Title: "Unknown Year Track"}), "[] Unknown Year Track"; got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRenderSanitizesSlashesInFieldValues(t *testing.T) {
	tpl, err := Parse("{artist}/{title}.{ext}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := tpl.Render(Fields{Artist: "AC/DC", Title: "T.N.T.", Ext: "mp3"})
	want := "AC-DC/T.N.T..mp3"
	if got != want {
		t.Errorf("Render() = %q, want %q (a slash in a field value must not add a path segment)", got, want)
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	if _, err := Parse("{nonsense}"); err == nil {
		t.Error("Parse with an unknown field: got nil error, want an error")
	}
}

func TestParseRejectsEmptyTemplate(t *testing.T) {
	if _, err := Parse("   "); err == nil {
		t.Error("Parse with a blank template: got nil error, want an error")
	}
}

func TestRenderAllStringFields(t *testing.T) {
	tpl, err := Parse("{artist}|{album_artist}|{album}|{title}|{format}|{bitrate_class}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := tpl.Render(Fields{
		Artist:       "Artist",
		AlbumArtist:  "Album Artist",
		Album:        "Album",
		Title:        "Title",
		Format:       "FLAC",
		BitrateClass: "lossless",
	})
	want := "Artist|Album Artist|Album|Title|FLAC|lossless"
	if got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}
