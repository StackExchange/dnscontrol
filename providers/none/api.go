package none

/*

A front-end to the SDK.

Use the Facade Pattern to create a simplified interface to the SDK.

This isn't needed if the SDK offers the functions you want.  Feel free to delete this file.

Rate-limiting and re-try logic should be here. This is the module that should
implement re-trying if the provider replies with a 429 or other "you're going
too fast" or "temporary error, please re-try later) errors.

Pagination logic.  This is the module that should implement any any pagination
logic.  The vendor's SDK might return data in pages (typically each reply
includes the next 100 records and you must request additional "pages" of
records.  This file is typically where logic to get all pages is implemented.

*/
