package converter

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// GUI-01: valid log → templated text, grouped by sender, "{time} - {text}" lines, LF endings.
func TestConvertValidExample(t *testing.T) {
	f, err := os.Open("../../testdata/example.xml")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()

	res, err := Convert(f)
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	want := "16/9/2010 - [i]Gabriella[/i]\n" +
		"22:45:43 - qual o nome da banda\n" +
		"22:45:48 - de metal\n" +
		"22:45:50 - que tu me fez ouvir?\n" +
		"\n" +
		"16/9/2010 - Ricardo\n" +
		"22:45:59 - rhapsody\n" +
		"22:46:03 - dragonforce\n" +
		"22:46:05 - e avantasia\n" +
		"\n" +
		"16/9/2010 - [i]Gabriella[/i]\n" +
		"22:46:13 - obrigada\n"

	if got := string(res.Content); got != want {
		t.Errorf("content mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}
	if strings.Contains(string(res.Content), "\r\n") {
		t.Errorf("content contains CRLF endings; want LF only")
	}
}

// GUI-01: message text stays verbatim — no HTML escaping of &, <, '.
func TestConvertVerbatimText(t *testing.T) {
	xmlIn := `<?xml version="1.0"?>
<Log FirstSessionID="1" LastSessionID="1">
<Message Date="16/9/2010" Time="22:45:43" DateTime="2010-09-17T01:45:43.093Z" SessionID="1">
<From><User FriendlyName="Ana"/></From><To><User FriendlyName="Bob"/></To>
<Text Style="">Tom &amp; Jerry &lt;3 it&#39;s fine</Text>
</Message>
</Log>`

	res, err := Convert(strings.NewReader(xmlIn))
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	want := "16/9/2010 - Ana\n22:45:43 - Tom & Jerry <3 it's fine\n"
	if got := string(res.Content); got != want {
		t.Errorf("verbatim text mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// GUI-02: malformed XML → error, never panic.
func TestConvertMalformedXML(t *testing.T) {
	_, err := Convert(strings.NewReader("<Log><Message Date="))
	if err == nil {
		t.Fatal("Convert(malformed) = nil error; want error")
	}
}

// GUI-02 / edge case: 0-byte input → error, never panic.
func TestConvertEmptyInput(t *testing.T) {
	_, err := Convert(strings.NewReader(""))
	if err == nil {
		t.Fatal("Convert(0-byte) = nil error; want error")
	}
}

// GUI-03: log with zero <Message> elements → ErrNoMessages.
func TestConvertZeroMessages(t *testing.T) {
	xmlIn := `<?xml version="1.0"?><Log FirstSessionID="1" LastSessionID="1"></Log>`
	_, err := Convert(strings.NewReader(xmlIn))
	if !errors.Is(err, ErrNoMessages) {
		t.Fatalf("Convert(zero messages) error = %v; want ErrNoMessages", err)
	}
}

// GUI-05: template is embedded — conversion works with the working directory
// moved away from the repo (no runtime template file read).
func TestConvertNoRuntimeTemplateRead(t *testing.T) {
	xmlIn := `<?xml version="1.0"?>
<Log FirstSessionID="1" LastSessionID="1">
<Message Date="16/9/2010" Time="22:45:43" DateTime="2010-09-17T01:45:43.093Z" SessionID="1">
<From><User FriendlyName="Ana"/></From><To><User FriendlyName="Bob"/></To>
<Text Style="">oi</Text>
</Message>
</Log>`

	t.Chdir(t.TempDir()) // no templates/ or *.tmpl reachable from cwd

	res, err := Convert(strings.NewReader(xmlIn))
	if err != nil {
		t.Fatalf("Convert with cwd outside repo returned error: %v", err)
	}
	want := "16/9/2010 - Ana\n22:45:43 - oi\n"
	if got := string(res.Content); got != want {
		t.Errorf("embedded-template output mismatch\ngot:  %q\nwant: %q", got, want)
	}
}
