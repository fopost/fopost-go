package fopost

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeParsesTheFormatsTheAPISends(t *testing.T) {
	cases := map[string]string{
		`"2026-09-01T10:00:00Z"`:      "2026-09-01T10:00:00Z",
		`"2026-09-01T10:00:00.000Z"`:  "2026-09-01T10:00:00Z",
		`"2026-09-01T10:00:00+02:00"`: "2026-09-01T08:00:00Z",
		`"2026-09-01"`:                "2026-09-01T00:00:00Z",
	}
	for input, want := range cases {
		var parsed Time
		if err := json.Unmarshal([]byte(input), &parsed); err != nil {
			t.Fatalf("Unmarshal(%s): %v", input, err)
		}
		if got := parsed.String(); got != want {
			t.Errorf("%s parsed to %s, want %s", input, got, want)
		}
	}
}

func TestTimeHandlesNullAndUnparseableValues(t *testing.T) {
	var null Time
	if err := json.Unmarshal([]byte("null"), &null); err != nil {
		t.Fatalf("Unmarshal(null): %v", err)
	}
	if !null.IsZero() || null.Raw != "" {
		t.Fatalf("null = %+v", null)
	}

	var odd Time
	if err := json.Unmarshal([]byte(`"whenever"`), &odd); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// Nothing is lost: the raw text survives even when it will not parse.
	if !odd.IsZero() || odd.Raw != "whenever" {
		t.Fatalf("odd = %+v", odd)
	}
	encoded, err := json.Marshal(odd)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(encoded) != `"whenever"` {
		t.Fatalf("re-encoded to %s", encoded)
	}
}

func TestTimeMarshalsAsRFC3339(t *testing.T) {
	encoded, err := json.Marshal(NewTime(time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(encoded) != `"2026-09-01T10:00:00Z"` {
		t.Fatalf("encoded = %s", encoded)
	}

	var zero Time
	encoded, err = json.Marshal(zero)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(encoded) != "null" {
		t.Fatalf("a zero time should encode as null, got %s", encoded)
	}
}

func TestTextAndThreadBuildContentBlocks(t *testing.T) {
	if blocks := Text("Hello"); len(blocks) != 1 || blocks[0].Text != "Hello" {
		t.Fatalf("Text = %+v", blocks)
	}
	blocks := Thread("one", "two", "three")
	if len(blocks) != 3 || blocks[2].Text != "three" {
		t.Fatalf("Thread = %+v", blocks)
	}
}

func TestPointerHelpers(t *testing.T) {
	if *Bool(true) != true || *String("x") != "x" || *Int(3) != 3 {
		t.Fatal("pointer helpers should carry their value")
	}
}
