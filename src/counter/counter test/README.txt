Not every function present in the non-test version of counter is present in the test
version.  There are a few functions that do minor reads, etc and return which is
difficult and probably unimportant to test using the golang testing package.

The primary purpose of the testing is to ensure the dumpfile is functioning correctly