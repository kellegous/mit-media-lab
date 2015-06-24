package mitml

import (
	"testing"
)

func assertParsedEmail(t *testing.T, txt, expEmail, expRest string) {
	email, rest, err := parseEmail(txt)
	if err != nil {
		t.Fatal(err)
	}

	if email != expEmail {
		t.Fatalf("expected email %s, got %s", expEmail, email)
	}

	if rest != expRest {
		t.Fatalf("expected rest %s, got %s", expRest, rest)
	}
}

func TestParseEmail(t *testing.T) {
	assertParsedEmail(t, "foo@bar.com", "foo@bar.com", "")
	assertParsedEmail(t, "foo@bar.com First Last", "foo@bar.com", "First Last")
	assertParsedEmail(t, "foo@bar.com      First   Last", "foo@bar.com", "First   Last")
	assertParsedEmail(t, "foo@bar.com      ", "foo@bar.com", "")

	if _, _, err := parseEmail("butter"); err == nil {
		t.Fatal("expected error for txt=\"butter\"")
	}

	if _, _, err := parseEmail("butter @@@"); err == nil {
		t.Fatal("expected error for txt=\"butter\"")
	}
}

func assertParsedNames(t *testing.T, txt, expFirst, expLast string) {
	first, last := parseNames(txt)
	if first != expFirst {
		t.Fatalf("expected first = %s, got %s (%s)", expFirst, first, txt)
	}
	if last != expLast {
		t.Fatalf("expected last = %s, got %s (%s)", expLast, last, txt)
	}
}

func TestParseNames(t *testing.T) {
	assertParsedNames(t, "Abe Lincoln", "Abe", "Lincoln")
	assertParsedNames(t, "Abe Lincoln Jr", "Abe", "Lincoln Jr")
	assertParsedNames(t, "Abe \"Lincoln Jr\"", "Abe", "Lincoln Jr")
	assertParsedNames(t, "Abe \"Lincoln   Jr\"", "Abe", "Lincoln   Jr")
	assertParsedNames(t, "Abe    Lincoln", "Abe", "Lincoln")
	assertParsedNames(t, "Abe\"Lincoln\"Jr   \"abelincolnjr \"", "AbeLincolnJr", "abelincolnjr ")
	assertParsedNames(t, "", "", "")
}
