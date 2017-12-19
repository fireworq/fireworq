package config

import (
	"os"
	"testing"
)

func TestGetSet(t *testing.T) {
	if Get("test1") != "" {
		t.Error("There must be no config value at first")
	}

	Set("test1", "foo")
	if Get("test1") != "foo" {
		t.Error("It should return stored config value")
	}

	Set("test1", "bar")
	if Get("test1") != "bar" {
		t.Error("It should return stored config value")
	}

	os.Setenv("FIREWORQ_TEST_FALLBACK_1", "env-value1")
	if Get("test_fallback_1") != "env-value1" {
		t.Error("It should fallback to environment")
	}

	SetDefault("test_fallback_2", "default-value2")
	if Get("test_fallback_2") != "default-value2" {
		t.Error("It should fallback to the default value")
	}

	os.Setenv("FIREWORQ_TEST_FALLBACK_3", "env-value3")
	SetDefault("test_fallback_3", "default-value3")
	if Get("test_fallback_3") != "env-value3" {
		t.Error("A value from environment variable should have higher precedence than the default value")
	}
}

func TestGetSetDefault(t *testing.T) {
	if GetDefault("default1") != "" {
		t.Error("There must be no config value at first")
	}

	SetDefault("default1", "bar")
	if GetDefault("default1") != "bar" {
		t.Error("It should return stored config value")
	}

	SetDefault("default1", "foo")
	if GetDefault("default1") != "foo" {
		t.Error("It should return stored config value")
	}

	os.Setenv("FIREWORQ_TEST_NO_FALLBACK_1", "env-value1")
	if GetDefault("test_no_fallback_1") != "" {
		t.Error("It should not fallback to environment")
	}
}

func TestLocally(t *testing.T) {
	original := Get("test_locally")

	Locally("test_locally", "some value", func() {
		if Get("test_locally") != "some value" {
			t.Error("Configuration value should be overridden in a block")
		}
	})
	if Get("test_locally") != original {
		t.Error("Configuration value should be restored outside the block")
	}

	Set("test_locally", "another value")

	Locally("test_locally", "some value", func() {
		if Get("test_locally") != "some value" {
			t.Error("Configuration value should be overridden in a block")
		}
	})
	if Get("test_locally") != "another value" {
		t.Error("Configuration value should be restored outside the block")
	}
}

func TestKeys(t *testing.T) {
	keys := Keys()
	if len(keys) <= 0 {
		t.Error("There must be some configuration keys")
	}
}
