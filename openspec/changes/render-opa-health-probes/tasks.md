# Tasks

- [ ] document the deployment health-probe gap and the listen-address-to-port design
- [ ] render `containerPort` from the normalized listen address in generated Deployment manifests
- [ ] add default readiness and liveness probes to the generated OPA container
- [ ] add regression tests for default and explicit listen-address render output
- [ ] update README and example output expectations to describe the new manifest behavior