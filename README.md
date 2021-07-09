# log

Simple logging library with functionality for enabling and disabling log levels
at runtime. A separate instance of `Logger` is intended for each package that
emit log messages. That way _FATAL_, _ERROR_, _WARNING_, _INFO_ and _DEBUG_
log messages can be enabled and disable for the specific packages.

*log*'s API is not yet stable.