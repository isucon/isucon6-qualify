package isucon6.web;

import org.apache.commons.lang3.ObjectUtils;

public class Config {

	static {

		host = ObjectUtils.defaultIfNull(System.getenv("ISUTAR_DB_HOST"), "localhost");
		port = ObjectUtils.defaultIfNull(System.getenv("ISUTAR_DB_PORT"), "3306");
		user = ObjectUtils.defaultIfNull(System.getenv("ISUTAR_DB_USER"), "isucon");
		password = ObjectUtils.defaultIfNull(System.getenv("ISUTAR_DB_PASSWORD"), "isucon");
		isudaOrigin = ObjectUtils.defaultIfNull(System.getenv("ISUDA_ORIGIN"), "http://localhost:5000");
		isutarOrigin = ObjectUtils.defaultIfNull(System.getenv("ISUTAR_ORIGIN"), "http://localhost:5001");
		isupamOrigin = ObjectUtils.defaultIfNull(System.getenv("ISUPAM_ORIGIN"), "http://localhost:5050");

		db = "isutar";
//		charset = "utf8mb4";
		charset = "utf8";
	}

	public static final String host;

	public static final String port;

	public static final String user;

	public static final String password;

	public static final String db;

	public static final String charset;

	public static final String isudaOrigin;

	public static final String isutarOrigin;

	public static final String isupamOrigin;
}
