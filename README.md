# IBM Security QRadar SOAR function runtime (golang)

This package enables the development and execution of custom IBM SOAR functions using Golang. While IBM officially recommends Python 3 for SOAR extensions, this project demonstrates an alternative that is more resource-efficient and unbounded by the Python thread limitations.

## Use cases
- Develop custom IBM SOAR functions in Golang instead of Python.
- Integrate existing Go libraries and tooling into SOAR playbooks.
- Build specialized utilities for SOAR administering and maintenance.

## Features & design principles
- STOMP Listener checks SOAR connectivity and credentials provided upon creation of struct.
- Dedicated listener per STOMP queue.
- Embeddable Function Logic: The runtime expects a user-defined function per listener. Manual dispatching is required is case of many functions per MD.

## Caveats
- Created upon reverse-engineered assumtions, there are no documentation for many aspects of internals of SOAR runtime and STOMP interactions;
- No input validation on runtime side;
- No SDK for function signature sync with SOAR;
- Provided as-is, not officially supported.

