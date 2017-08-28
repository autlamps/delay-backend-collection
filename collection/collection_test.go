package collection

import "testing"

func TestCreateQueryURL(t *testing.T) {

	tests := []struct {
		values []string
		key    string
		env    Env
		out    string
	}{
		{[]string{"1", "2", "3", "4"}, "vehicleid", Env{ApiKey: "1234"}, "api_key=1234&vehicleid=1%2C2%2C3%2C4%2C"},
	}

	for _, test := range tests {

		output := test.env.createQueryURL(test.key, test.values)

		if output != test.out {
			t.Errorf("Test failed. Expected %v, got %v", test.out, output)
		}
	}
}
