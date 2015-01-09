package httpclient

httpclient.
    Defaults(httpclient.Config {
        Before: xx
    }).

// ==== Begin Request    
    Begin().

// ==== Build Request begin
    Header("a", "b").
    Cookie("c", "d").
    Field("a", "b").
    Query("a", "b").

    Data(FormFile("/path/to/file")).
    Data(NewFormFile("/"))
    File("ff").
    FormFile("key", "/test.txt").
    File("/test.txt").

// ==== Send Request
    Get("url").
    Post("url").
    // PostJSON("url").
    Put("url").
    Delete("url").
    Head("url").
    Trace("url").
    Connect("url").

// Hook

    Hook(func(client, builder))
