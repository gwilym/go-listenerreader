# listenerreader

listenerreader provides a means of treating a net.Listener as a single io.Reader by tokenizing traffic received over the connections and merging them into one readable stream.

So far this is just a proof of concept and is not considered stable for production use.
