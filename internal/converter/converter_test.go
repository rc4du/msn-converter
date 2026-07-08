package converter

import (
	"errors"
	"os"
	"path/filepath"
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

// GUI-04: filename derived from the first message: {date /→_}_{time}_{receiver}.txt,
// Windows-illegal chars → "-". Example fixture per spec.
func TestConvertFileNameExample(t *testing.T) {
	f, err := os.Open("../../testdata/example.xml")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()

	res, err := Convert(f)
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if want := "16_9_2010_22-45-43_Ricardo.txt"; res.FileName != want {
		t.Errorf("FileName = %q; want %q", res.FileName, want)
	}
}

// GUI-04: every Windows-illegal char class (\ / : * ? " < > |) in the receiver
// name is replaced by "-"; receiver with "/" ([i]Gabriella[/i]) included.
func TestConvertFileNameSanitization(t *testing.T) {
	cases := []struct {
		name     string
		receiver string // raw FriendlyName (XML-escaped when templated below)
		want     string
	}{
		{"slash-receiver", "[i]Gabriella[/i]", "16_9_2010_22-45-43_[i]Gabriella[-i].txt"},
		{"backslash", `a\b`, "16_9_2010_22-45-43_a-b.txt"},
		{"slash", "a/b", "16_9_2010_22-45-43_a-b.txt"},
		{"colon", "a:b", "16_9_2010_22-45-43_a-b.txt"},
		{"asterisk", "a*b", "16_9_2010_22-45-43_a-b.txt"},
		{"question", "a?b", "16_9_2010_22-45-43_a-b.txt"},
		{"quote", `a&quot;b`, "16_9_2010_22-45-43_a-b.txt"},
		{"less-than", "a&lt;b", "16_9_2010_22-45-43_a-b.txt"},
		{"greater-than", "a&gt;b", "16_9_2010_22-45-43_a-b.txt"},
		{"pipe", "a|b", "16_9_2010_22-45-43_a-b.txt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xmlIn := `<?xml version="1.0"?>
<Log FirstSessionID="1" LastSessionID="1">
<Message Date="16/9/2010" Time="22:45:43" DateTime="2010-09-17T01:45:43.093Z" SessionID="1">
<From><User FriendlyName="Ana"/></From><To><User FriendlyName="` + tc.receiver + `"/></To>
<Text Style="">oi</Text>
</Message>
</Log>`
			res, err := Convert(strings.NewReader(xmlIn))
			if err != nil {
				t.Fatalf("Convert returned error: %v", err)
			}
			if res.FileName != tc.want {
				t.Errorf("FileName = %q; want %q", res.FileName, tc.want)
			}
		})
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

// GUI-14 (mechanics): ConvertFile writes the derived file into outDir with
// content identical to Convert's output, and returns the written name.
func TestConvertFileWritesOutput(t *testing.T) {
	outDir := t.TempDir()

	name, err := ConvertFile("../../testdata/example.xml", outDir)
	if err != nil {
		t.Fatalf("ConvertFile returned error: %v", err)
	}
	if want := "16_9_2010_22-45-43_Ricardo.txt"; name != want {
		t.Errorf("returned name = %q; want %q", name, want)
	}

	got, err := os.ReadFile(filepath.Join(outDir, name))
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	f, err := os.Open("../../testdata/example.xml")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer f.Close()
	res, err := Convert(f)
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if string(got) != string(res.Content) {
		t.Errorf("written content differs from Convert output\ngot:  %q\nwant: %q", got, res.Content)
	}
}

// GUI-14: pre-existing file with the same name is overwritten (old content gone).
func TestConvertFileOverwritesExisting(t *testing.T) {
	outDir := t.TempDir()
	target := filepath.Join(outDir, "16_9_2010_22-45-43_Ricardo.txt")
	if err := os.WriteFile(target, []byte("OLD CONTENT MUCH LONGER THAN THE REAL OUTPUT SO TRUNCATION MATTERS "+strings.Repeat("x", 4096)), 0o644); err != nil {
		t.Fatalf("seeding pre-existing file: %v", err)
	}

	if _, err := ConvertFile("../../testdata/example.xml", outDir); err != nil {
		t.Fatalf("ConvertFile returned error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading overwritten file: %v", err)
	}
	if strings.Contains(string(got), "OLD CONTENT") {
		t.Errorf("old content still present after overwrite")
	}
	if !strings.HasPrefix(string(got), "16/9/2010 - [i]Gabriella[/i]\n") {
		t.Errorf("overwritten file does not start with expected converted output; got %q", string(got)[:min(len(got), 60)])
	}
}

// ConvertFile: unreadable input path → error, no panic.
func TestConvertFileUnreadableInput(t *testing.T) {
	_, err := ConvertFile(filepath.Join(t.TempDir(), "does-not-exist.xml"), t.TempDir())
	if err == nil {
		t.Fatal("ConvertFile(nonexistent input) = nil error; want error")
	}
}

// ConvertFile: nonexistent output dir → error, no panic.
func TestConvertFileBadOutputDir(t *testing.T) {
	_, err := ConvertFile("../../testdata/example.xml", filepath.Join(t.TempDir(), "missing-subdir"))
	if err == nil {
		t.Fatal("ConvertFile(nonexistent outDir) = nil error; want error")
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
