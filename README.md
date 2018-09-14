# Validating Admission Controller
This go controller takes care of `accepting` and `dropping` the change in annotations of any Openshift application depending upon the `whitelist` that is allowed to be exposed to CERN `technical network`.

### Logic behind the Controller

```
If route has annotation `router.cern.ch/technical-network-access: true` then
  Check if route annotation `router.cern.ch/network-visibility` has value `Internet`: 
    in this case route is requesting visibility on TN+Internet+Intranet; 
    otherwise it is requesting TN+Intranet
  Check value of label `router.cern.ch/technical-network-allowed` on the parent namespace: 
    If `Intranet`, the project is allowed to expose routes to TN+Intranet; 
    if `Internet`, the project is allowed to expose routes to TN+Intranet and TN+Internet+Intranet
    if not present or other value: project is not allowed to expose routes to TN at all
  If the requested route visibility is not allowed by the label on the namespace, reject the route creation/modification (otherwise, accept the change)
```

1. TN : Technical Network
2. Openshift's conventions on standard Openshift route annotations are followed, so the following spellings in annotation/label values are equivalent: Internet, internet, INTERNET.
3. Similarly, True/true/TRUE are equal.

### Deployment
