// Default configuration
DefaultProvider: "togetherai"

// Environment-specific configurations using switch pattern
MaxTokens: [
    if #Meta.Environment.Type == "development" {100},
    if #Meta.Environment.Type == "test" {50},
    150, // default case
][0]

Temperature: [
    if #Meta.Environment.Type == "development" {0.5},
    if #Meta.Environment.Type == "test" {0.0},
    0.7, // default case
][0]
