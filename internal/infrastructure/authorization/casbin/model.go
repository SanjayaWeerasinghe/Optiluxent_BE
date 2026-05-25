package casbin

// RBACModel is the Casbin model configuration for role-based access control.
// p = policy: sub (role), obj (resource), act (action)
// g = role inheritance: user, role
// e = effect combinator: any allow wins
// m = matcher: role match OR super_admin bypass
const RBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.sub == "super_admin"
`
