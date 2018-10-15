package tests

import "testing"

func subTestSearch(t *testing.T, mc *mockServer) {
	runStep(t, mc, "KNN", keys_KNN_test)
	runStep(t, mc, "WITHIN_CIRCLE", keys_WITHIN_CIRCLE_test)
	runStep(t, mc, "INTERSECTS_CIRCLE", keys_INTERSECTS_CIRCLE_test)
}

func keys_KNN_test(mc *mockServer) error {
	return mc.DoBatch([][]interface{}{
		{"SET", "mykey", "1", "POINT", 5, 5}, {"OK"},
		{"SET", "mykey", "2", "POINT", 19, 19}, {"OK"},
		{"SET", "mykey", "3", "POINT", 12, 19}, {"OK"},
		{"SET", "mykey", "4", "POINT", -5, 5}, {"OK"},
		{"SET", "mykey", "5", "POINT", 33, 21}, {"OK"},
		{"NEARBY", "mykey", "LIMIT", 10, "DISTANCE", "POINTS", "POINT", 20, 20}, {
			"" +
				"[0 [" +
				("" +
					"[2 [19 19] 152808.67164036975] " +
					"[3 [12 19] 895945.1409106685] " +
					"[5 [33 21] 1448929.5916252395] " +
					"[1 [5 5] 2327116.1069888202] " +
					"[4 [-5 5] 3227402.6159841116]") +
				"]]"},
	})
}

func keys_WITHIN_CIRCLE_test(mc *mockServer) error {
	return mc.DoBatch([][]interface{}{
		{"SET", "mykey", "1", "POINT", 37.7335, -122.4412}, {"OK"},
		{"SET", "mykey", "2", "POINT", 37.7335, -122.44121}, {"OK"},
		{"SET", "mykey", "3", "OBJECT", `{"type":"LineString","coordinates":[[-122.4408378,37.7341129],[-122.4408378,37.733]]}`}, {"OK"},
		{"SET", "mykey", "4", "OBJECT", `{"type":"Polygon","coordinates":[[[-122.4408378,37.7341129],[-122.4408378,37.733],[-122.44,37.733],[-122.44,37.7341129],[-122.4408378,37.7341129]]]}`}, {"OK"},
		{"SET", "mykey", "5", "OBJECT", `{"type":"MultiPolygon","coordinates":[[[[-122.4408378,37.7341129],[-122.4408378,37.733],[-122.44,37.733],[-122.44,37.7341129],[-122.4408378,37.7341129]]]]}`}, {"OK"},
		{"SET", "mykey", "6", "POINT", -5, 5}, {"OK"},
		{"SET", "mykey", "7", "POINT", 33, 21}, {"OK"},
		{"WITHIN", "mykey", "IDS", "CIRCLE", 37.7335, -122.4412, 1000}, {
			"[0 [1 2 3 4 5]]"},
		{"WITHIN", "mykey", "IDS", "CIRCLE", 37.7335, -122.4412, 10}, {
			"[0 [1 2]]"},
	})
}

func keys_INTERSECTS_CIRCLE_test(mc *mockServer) error {
	return mc.DoBatch([][]interface{}{
		{"SET", "mykey", "1", "POINT", 37.7335, -122.4412}, {"OK"},
		{"SET", "mykey", "2", "POINT", 37.7335, -122.44121}, {"OK"},
		{"SET", "mykey", "3", "OBJECT", `{"type":"LineString","coordinates":[[-122.4408378,37.7341129],[-122.4408378,37.733]]}`}, {"OK"},
		{"SET", "mykey", "4", "OBJECT", `{"type":"Polygon","coordinates":[[[-122.4408378,37.7341129],[-122.4408378,37.733],[-122.44,37.733],[-122.44,37.7341129],[-122.4408378,37.7341129]]]}`}, {"OK"},
		{"SET", "mykey", "5", "OBJECT", `{"type":"MultiPolygon","coordinates":[[[[-122.4408378,37.7341129],[-122.4408378,37.733],[-122.44,37.733],[-122.44,37.7341129],[-122.4408378,37.7341129]]]]}`}, {"OK"},
		{"SET", "mykey", "6", "POINT", -5, 5}, {"OK"},
		{"SET", "mykey", "7", "POINT", 33, 21}, {"OK"},
		{"INTERSECTS", "mykey", "IDS", "CIRCLE", 37.7335, -122.4412, 70}, {
			"[0 [1 2 3 4 5]]"},
		{"INTERSECTS", "mykey", "IDS", "CIRCLE", 37.7335, -122.4412, 10}, {
			"[0 [1 2]]"},
	})
}
