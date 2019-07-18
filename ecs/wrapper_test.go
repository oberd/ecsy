package ecs

import "testing"

var commandOverrideTests = []struct {
	commandString    string
	expectedOverride []string
}{
	{"node src/bin/cli.js generate-snapshots", []string{"node", "src/bin/cli.js", "generate-snapshots"}},
	{
		"sh -c \"/var/www/qc2-crons; php /var/www/app/Console/cake.php cron generateScorecardsData\"",
		[]string{"sh", "-c", "/var/www/qc2-crons; php /var/www/app/Console/cake.php cron generateScorecardsData"},
	},
}

func TestCommandParsing(t *testing.T) {
	for _, args := range commandOverrideTests {
		t.Run(args.commandString, func(t *testing.T) {
			override, err := parseCommandOverride(args.commandString)
			if err != nil && args.expectedOverride != nil {
				t.Errorf("expected error")
			}
			for i := range override {
				if override[i] != args.expectedOverride[i] {
					t.Errorf("expected override %v , got %v", args.expectedOverride, override)
				}
			}
		})
	}
}
