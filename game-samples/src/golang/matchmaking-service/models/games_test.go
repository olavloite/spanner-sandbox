package models

import "testing"

func TestDetermineWinnerWithPlayers(t *testing.T) {
	var tests = []string{"ea32ff20-e10f-42c4-80d1-e0e1970eeb56", "3349f46a-215d-42e9-ab3a-759883cfeb2e",
		"7dacee48-9380-439d-995b-86d811316c14", "5a6bd9fa-6676-47b9-97a8-cc88a940b3d3",
		"a1781798-7fe4-4b32-a376-071153486357", "cf0e0641-9abf-4a52-a36c-b0278867dba9",
		"a4f13a8f-dc03-4704-8f2d-c6c26e0479c0", "d4a63a17-4b3c-4c80-9fa2-7e8ad14d2996",
		"ee3d969e-b47f-449c-9836-b766148959fe", "aa8b8b57-0ecf-49e3-84ab-4117dd86eff4",
		"2e5eb94d-bb4e-40e5-a48a-3281b7db8718", "57328c25-9810-4734-a002-87dab340c965",
		"3e937267-e579-4282-b1ee-6dba915b475e", "15315397-af21-4be6-b824-5740f85ba273"}

	res := determineWinner(tests)

	if res == "" {
		t.Error(
			"For", tests,
			"expected", "not empty string",
			"got", "empty string",
		)
	}
}

func TestDetermineWinnerWithoutPlayers(t *testing.T) {
	var tests = []string{}

	res := determineWinner(tests)

	if res != "" {
		t.Error(
			"For", tests,
			"expected", "empty string",
			"got", "not empty string",
		)
	}
}
