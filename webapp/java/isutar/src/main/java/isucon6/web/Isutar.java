package isucon6.web;

import static spark.Spark.*;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLEncoder;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.ResultSetMetaData;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.apache.commons.lang3.StringUtils;

import net.arnx.jsonic.JSON;
import spark.Request;
import spark.Route;

public class Isutar {

	public static void main(String[] args) throws Exception {

		port(5001);

		before("*", (request, response) -> {
			Connection connection = getConnection(request);
			request.attribute("connection", connection);
		});

		get("/initialize", initialize);
		get("/stars", getStars);
		post("/stars", postStars);

		after("*", (request, response) -> {
			Connection connection = (Connection) request.attribute("connection");
			connection.commit();
			connection.close();
		});

		exception(Exception.class, (exception, request, response) -> {

			try {
				Connection connection = (Connection) request.attribute("connection");exception.printStackTrace();
				connection.rollback();
				connection.close();
			} catch (SQLException e) {
				throw new RuntimeException("SystemException", e);
			}
		});
	}

	protected static Route initialize = (request, response) -> {

		Connection connection = getConnection(request);

		execute(connection, "TRUNCATE star");

		response.type("text/json");
		return "{status: \"ok\"}";
	};

	protected static Route getStars = (request, response) -> {

		Connection connection = getConnection(request);

		List<Map<String, Object>> stars = select(connection, "SELECT * FROM star WHERE keyword = ?",
				request.queryParams("keyword"));

		Map<String, Object> result = new HashMap<>();
		result.put("stars", stars);

		response.type("text/json");
		return JSON.encode(result);
	};

	protected static Route postStars = (request, response) -> {

		Connection connection = getConnection(request);

		String keyword = request.queryParams("keyword");
		if (StringUtils.isEmpty(keyword)) {
			keyword = request.params("keyword");
		}

		Map<String, String> httpResult = httpGet(
				Config.isudaOrigin + "/keyword/" + urlEncode(keyword));
		if ("404".equals(httpResult.get("status"))) {
			halt(404);
		} else if (!"200".equals(httpResult.get("status"))) {
			throw new RuntimeException("http request error");
		}

		String user = request.queryParams("user");
		if (StringUtils.isEmpty(user)) {
			user = request.params("user");
		}

		execute(connection, "INSERT INTO star (keyword, user_name, created_at) VALUES (?, ?, NOW())", keyword, user);

		response.type("text/json");
		return "{result: \"ok\"}";
	};

	private static List<Map<String, Object>> select(Connection connection, String sql, Object... params)
			throws SQLException {
		//
		try (PreparedStatement statement = connection.prepareStatement(sql)) {

			setParams(statement, params);

			try (ResultSet rs = statement.executeQuery()) {

				ResultSetMetaData metaData = rs.getMetaData();
				int columnCount = metaData.getColumnCount();

				List<Map<String, Object>> result = new ArrayList<>();

				while (rs.next()) {

					Map<String, Object> item = convertToMap(rs, metaData, columnCount);

					result.add(item);
				}

				return result;
			}
		}
	}

	private static int execute(Connection connection, String sql) throws SQLException {
		//
		return execute(connection, sql, new Object[0]);
	}

	private static int execute(Connection connection, String sql, Object... params) throws SQLException {
		//
		try (PreparedStatement statement = connection.prepareStatement(sql)) {

			setParams(statement, params);

			return statement.executeUpdate();
		}
	}

	private static void setParams(PreparedStatement statement, Object... params) throws SQLException {
		//
		for (int i = 0; i < params.length; i++) {
			if (params[i] instanceof Integer) {
				statement.setInt(i + 1, (Integer) params[i]);
			} else {
				statement.setString(i + 1, (String) params[i]);
			}
		}
	}

	private static Map<String, Object> convertToMap(ResultSet rs, ResultSetMetaData metaData, int columnCount)
			throws SQLException {
		Map<String, Object> item = new HashMap<>();

		for (int i = 0; i < columnCount; i++) {
			item.put(metaData.getColumnName(i + 1), rs.getObject(i + 1));
		}
		return item;
	}

	private static Connection getConnection(Request request) {

		if (request.attribute("connection") != null) {
			return request.attribute("connection");
		}

		//
		try {
			//
			Class.forName("com.mysql.jdbc.Driver");

			String serverName = Config.host;
			String port = Config.port;
			String databaseName = Config.db;
			String user = Config.user;
			String password = Config.password;

			String url = "jdbc:mysql://" + serverName + ":" + port + "/" + databaseName
					+ "?useUnicode=true&characterEncoding=" + Config.charset;

			Connection connection = DriverManager.getConnection(url, user, password);
			connection.setAutoCommit(false);

			try (Statement statement = connection.createStatement()) {
				//
				statement.execute("SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'");
				statement.execute("SET NAMES utf8mb4");
			}

			request.attribute("connection", connection);

			return connection;

		} catch (ClassNotFoundException | SQLException e) {
			//
			throw new RuntimeException("SystemException", e);
		}
	}

	private static Map<String, String> httpGet(String urlString) {

		Map<String, String> result = new HashMap<>();

		HttpURLConnection con = null;

		try {

			URL url = new URL(urlString);

			con = (HttpURLConnection) url.openConnection();
			con.setRequestMethod("GET");

			result.put("status", String.valueOf(con.getResponseCode()));
			if (con.getResponseCode() != HttpURLConnection.HTTP_OK) {
				//
				return result;
			}

			try (BufferedReader in = new BufferedReader(new InputStreamReader(con.getInputStream(), "UTF-8"))) {

				StringBuilder body = new StringBuilder();
				String line = null;
				while ((line = in.readLine()) != null) {
					body.append(line);
				}

				result.put("body", body.toString());

				return result;
			}

		} catch (IOException e) {
			throw new RuntimeException("SystemException", e);
		}
	}

	private static String urlEncode(String value) {

		try {
			String result = URLEncoder.encode(value, "UTF-8");
			result = result.replace("+", "%20");

			return result;

		} catch (Exception e) {
			throw new RuntimeException("SystemException", e);
		}
	}
}
