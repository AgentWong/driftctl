package acceptance

import (
	"os"
	"reflect"
	"testing"
)

func TestAccTestCase_resolveTerraformEnv(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ACC_TEST_VAR", "foobar")
	_ = os.Setenv("TEST_VAR", "barfoo")
	_ = os.Setenv("TEST_VAR_2", "barfoo")
	_ = os.Setenv("ACC_TEST_VAR_3", "")
	_ = os.Setenv("TEST_VAR_3", "barfoo")
	_ = os.Setenv("TEST_VAR_4", "barfoo")
	_ = os.Setenv("ACC_TEST_VAR_4", "")

	testCase := AccTestCase{}
	env := testCase.resolveTerraformEnv()
	expected := map[string]string{
		"TEST_VAR":   "foobar",
		"TEST_VAR_2": "barfoo",
		"TEST_VAR_3": "",
		"TEST_VAR_4": "",
	}

	if !reflect.DeepEqual(expected, env) {
		t.Fatalf("Variable env override not working, got: %+v, expected %+v", env, expected)
	}
}
