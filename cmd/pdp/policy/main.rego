package policy

import future.keywords.if
import future.keywords.in

default allow := false

allow if {
    data.policy.rbac.allow
}
