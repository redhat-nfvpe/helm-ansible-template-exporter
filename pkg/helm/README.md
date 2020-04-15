# helm

Helm is a go wrapper package which is meant to aid in the conversion of Helm Templates to Ansible Playbook Role(s).
There are two main pieces of functionality included in this internal package:

1) convert.go
2) helmfuncs.go

# convert.go

Contains the utility functions to convert from Helm -> Ansible Playbook Role.

# helmfuncs.go

Helm abstracts custom Template functions in its parser (namely define, include, & Sprig functions), yet doesn't make the
funcMap public.  For more information see the
[Helm Engine funcMap abstraction](https://github.com/helm/helm/blob/master/pkg/engine/funcs.go#L44).

That means that the functionality they add, which is quite useful, is not readily extendable through a Go client, since
a normal Go client doesn't have the visibility to call non-exported module functions.  Although this can be done through
reflection (and probably should be done this way at some point), it was much easier to just copy in the functionality
for now.  Since the license is Apache License 2.0, and since we did not alter any copyright, this should be fine.

This file was modified slightly only to change the package name, and to export the "HelmFuncMap" function for use in
other files.
