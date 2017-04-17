package viper

import (
	"fmt"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

//stubs for PFlag Values
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string {
	return fmt.Sprintf("%s", *s)
}

func TestBindFlagValueSet(t *testing.T) {
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)

	var testValues = map[string]*string{
		"host":     nil,
		"port":     nil,
		"endpoint": nil,
	}

	var mutatedTestValues = map[string]string{
		"host":     "localhost",
		"port":     "6060",
		"endpoint": "/public",
	}

	for name := range testValues {
		testValues[name] = flagSet.String(name, "", "test")
	}

	flagValueSet := pflagValueSet{flagSet}

	err := BindFlagValues(flagValueSet)
	if err != nil {
		t.Fatalf("error binding flag set, %v", err)
	}

	flagSet.VisitAll(func(flag *pflag.Flag) {
		flag.Value.Set(mutatedTestValues[flag.Name])
		flag.Changed = true
	})

	for name, expected := range mutatedTestValues {
		assert.Equal(t, Get(name), expected)
	}
}

func TestBindFlagValue(t *testing.T) {
	var testString = "testing"
	var testValue = newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "testflag",
		Value:   testValue,
		Changed: false,
	}

	flagValue := pflagValue{flag}
	BindFlagValue("testvalue", flagValue)

	assert.Equal(t, testString, Get("testvalue"))

	flag.Value.Set("testing_mutate")
	flag.Changed = true //hack for pflag usage

	assert.Equal(t, "testing_mutate", Get("testvalue"))
}

func TestBindPFlags(t *testing.T) {
	v := New() // create independent Viper object
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)

	var testValues = map[string]*string{
		"host":     nil,
		"port":     nil,
		"endpoint": nil,
	}

	var mutatedTestValues = map[string]string{
		"host":     "localhost",
		"port":     "6060",
		"endpoint": "/public",
	}

	for name := range testValues {
		testValues[name] = flagSet.String(name, "", "test")
	}

	err := v.BindPFlags(flagSet)
	if err != nil {
		t.Fatalf("error binding flag set, %v", err)
	}

	flagSet.VisitAll(func(flag *pflag.Flag) {
		flag.Value.Set(mutatedTestValues[flag.Name])
		flag.Changed = true
	})

	for name, expected := range mutatedTestValues {
		assert.Equal(t, expected, v.Get(name))
	}

}

func TestBindPFlagsStringSlice(t *testing.T) {
	tests := []struct {
		Expected []string
		Value    string
	}{
		{[]string{}, ""},
		{[]string{"jeden"}, "jeden"},
		{[]string{"dwa", "trzy"}, "dwa,trzy"},
		{[]string{"cztery", "piec , szesc"}, "cztery,\"piec , szesc\""},
	}

	v := New() // create independent Viper object
	defaultVal := []string{"default"}
	v.SetDefault("stringslice", defaultVal)

	for _, testValue := range tests {
		flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flagSet.StringSlice("stringslice", testValue.Expected, "test")

		for _, changed := range []bool{true, false} {
			flagSet.VisitAll(func(f *pflag.Flag) {
				f.Value.Set(testValue.Value)
				f.Changed = changed
			})

			err := v.BindPFlags(flagSet)
			if err != nil {
				t.Fatalf("error binding flag set, %v", err)
			}

			type TestStr struct {
				StringSlice []string
			}
			val := &TestStr{}
			if err := v.Unmarshal(val); err != nil {
				t.Fatalf("%+#v cannot unmarshal: %s", testValue.Value, err)
			}
			if changed {
				assert.Equal(t, testValue.Expected, val.StringSlice)
			} else {
				assert.Equal(t, defaultVal, val.StringSlice)
			}
		}
	}
}

func TestBindPFlag(t *testing.T) {
	var testString = "testing"
	var testValue = newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "testflag",
		Value:   testValue,
		Changed: false,
	}

	BindPFlag("testvalue", flag)

	assert.Equal(t, testString, Get("testvalue"))

	flag.Value.Set("testing_mutate")
	flag.Changed = true //hack for pflag usage

	assert.Equal(t, "testing_mutate", Get("testvalue"))

}

func TestBindPFlagCaseInsensivity(t *testing.T) {
	v := New()

	var testString = "green"
	var testValue = newStringValue(testString, &testString)

	flag := &pflag.Flag{
		Name:    "eyeballs",
		Value:   testValue,
		Changed: true,
	}

	v.BindPFlag("eYEs", flag)
	assert.Equal(t, "green", v.Get("eyes"))
}
