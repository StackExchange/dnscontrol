package models

// func Test_UpdateSplitHorizonNames(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		dc       *DomainConfig
// 		expected *DomainConfig
// 	}{
// 		{
// 			name: "testNoTag",
// 			dc: &DomainConfig{
// 				Name: "example.com",
// 			},
// 			expected: &DomainConfig{
// 				Name: "example.com",
// 				Metadata: map[string]string{
// 					DomainUniqueName: "example.com",
// 					DomainTag:        "",
// 				},
// 			},
// 		},
// 		{
// 			name: "testEmptyTag",
// 			dc: &DomainConfig{
// 				Name: "example.com!",
// 			},
// 			expected: &DomainConfig{
// 				Name: "example.com",
// 				Metadata: map[string]string{
// 					DomainUniqueName: "example.com",
// 					DomainTag:        "",
// 				},
// 			},
// 		},
// 		{
// 			name: "testWithTag",
// 			dc: &DomainConfig{
// 				Name: "example.com!john",
// 			},
// 			expected: &DomainConfig{
// 				Name: "example.com",
// 				Metadata: map[string]string{
// 					DomainUniqueName: "example.com!john",
// 					DomainTag:        "john",
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			//tt.dc.UpdateSplitHorizonNames()
// 			if tt.dc.Name != tt.expected.Name {
// 				t.Errorf("expected name %s, got %s", tt.expected.Name, tt.dc.Name)
// 			}
// 			if tt.dc.Metadata[DomainUniqueName] != tt.expected.Metadata[DomainUniqueName] {
// 				t.Errorf("expected unique name %s, got %s", tt.expected.Metadata[DomainUniqueName], tt.dc.Metadata[DomainUniqueName])
// 			}
// 			if tt.dc.Metadata[DomainTag] != tt.expected.Metadata[DomainTag] {
// 				t.Errorf("expected tag %s, got %s", tt.expected.Metadata[DomainTag], tt.dc.Metadata[DomainTag])
// 			}
// 		})
// 	}
// }
