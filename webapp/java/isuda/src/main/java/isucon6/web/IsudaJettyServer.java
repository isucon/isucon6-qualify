package isucon6.web;

import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.util.thread.QueuedThreadPool;
import org.eclipse.jetty.util.thread.ThreadPool;

import spark.embeddedserver.jetty.JettyServerFactory;

public class IsudaJettyServer implements JettyServerFactory {

	/**
	 * Creates a Jetty server.
	 *
	 * @param maxThreads
	 *            maxThreads
	 * @param minThreads
	 *            minThreads
	 * @param threadTimeoutMillis
	 *            threadTimeoutMillis
	 * @return a new jetty server instance
	 */
	public Server create(int maxThreads, int minThreads, int threadTimeoutMillis) {
		Server server;

		if (maxThreads > 0) {
			int max = maxThreads;
			int min = (minThreads > 0) ? minThreads : 8;
			int idleTimeout = (threadTimeoutMillis > 0) ? threadTimeoutMillis : 60000;

			server = new Server(new QueuedThreadPool(max, min, idleTimeout));
		} else {
			server = new Server();
		}

		server.setAttribute("org.eclipse.jetty.server.Request.maxFormContentSize", 5000000);

		return server;
	}

	/**
	 * Creates a Jetty server with supplied thread pool
	 *
	 * @param threadPool
	 *            thread pool
	 * @return a new jetty server instance
	 */
	@Override
	public Server create(ThreadPool threadPool) {
		return threadPool != null ? new Server(threadPool) : new Server();
	}

}
