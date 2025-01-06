package policy

default allow = false

allow = true if {
	input.method == "GET"
	input.path == "/allowed"
}
