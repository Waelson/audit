import React, { useEffect, useState } from "react";
import axios from "axios";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { coy } from "react-syntax-highlighter/dist/esm/styles/prism";

const App = () => {
  const [filters, setFilters] = useState([]);
  const [query, setQuery] = useState({
    application: "",
    db_name: "",
    db_schema: "",
    db_table: "",
    event_operation: "",
    start_date: "",
    end_date: "",
  });
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [expandedRow, setExpandedRow] = useState(null);

  // Fetch filters on page load
  useEffect(() => {
    const fetchFilters = async () => {
      try {
        const response = await axios.get("http://localhost:5050/api/filters");
        setFilters(response.data);
      } catch (error) {
        console.error("Failed to fetch filters", error);
      }
    };
    fetchFilters();
  }, []);

  // Handle input changes
  const handleChange = (e) => {
    const { name, value } = e.target;
    setQuery((prevQuery) => ({ ...prevQuery, [name]: value }));
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      const response = await axios.get("http://localhost:5050/api/audit-trail", {
        params: query,
      });
      setResults(response.data);
    } catch (error) {
      console.error("Failed to fetch audit trails", error);
    } finally {
      setLoading(false);
    }
  };

  const getDatabases = (application) => {
    const appData = filters.find((f) => f.application === application);
    return appData ? appData.databases : [];
  };

  const getSchemas = (application, dbName) => {
    const databases = getDatabases(application);
    const db = databases.find((d) => d.name === dbName);
    return db ? db.schemas : [];
  };

  const getTables = (application, dbName, schemaName) => {
    const schemas = getSchemas(application, dbName);
    const schema = schemas.find((s) => s.name === schemaName);
    return schema ? schema.tables : [];
  };

  const toggleRow = (index) => {
    setExpandedRow(expandedRow === index ? null : index);
  };

  const getOperationLabel = (operation) => {
    switch (operation) {
      case "c":
        return "Create";
      case "u":
        return "Update";
      case "d":
        return "Delete";
      default:
        return operation; // Retorna o valor original caso não seja "C", "U" ou "D"
    }
  };

  const styles = {
    container: {
      fontFamily: "Arial, sans-serif",
      padding: "20px",
      maxWidth: "auto",
      margin: "0 auto",
    },
    header: {
      textAlign: "center",
      marginBottom: "20px",
      color: "#333",
    },
    form: {
      display: "grid",
      gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
      gap: "15px",
      marginBottom: "20px",
    },
    formGroup: {
      display: "flex",
      flexDirection: "column",
    },
    label: {
      marginBottom: "5px",
      fontWeight: "bold",
      color: "#555",
    },
    select: {
      padding: "8px",
      border: "1px solid #ccc",
      borderRadius: "5px",
    },
    input: {
      padding: "8px",
      border: "1px solid #ccc",
      borderRadius: "5px",
    },
    button: {
      gridColumn: "span 2",
      padding: "10px 20px",
      backgroundColor: "#4CAF50",
      color: "white",
      border: "none",
      borderRadius: "5px",
      cursor: "pointer",
      textAlign: "center",
    },
    buttonDisabled: {
      backgroundColor: "#ccc",
      cursor: "not-allowed",
    },
    tableContainer: {
      marginTop: "20px",
    },
    table: {
      width: "100%",
      borderCollapse: "collapse",
    },
    th: {
      borderBottom: "2px solid #ddd",
      textAlign: "left",
      padding: "10px",
      backgroundColor: "#f9f9f9",
    },
    td: {
      borderBottom: "1px solid #ddd",
      textAlign: "left",
      padding: "10px",
    },
    noResults: {
      textAlign: "center",
      color: "#888",
    },
    expandedRow: {
      backgroundColor: "#f9f9f9",
      padding: "10px",
    },
  };

  return (
      <div style={styles.container}>
        <h1 style={styles.header}>Audit Trail Query</h1>

        <form onSubmit={handleSubmit} style={styles.form}>
          <div style={styles.formGroup}>
            <label style={styles.label}>Application</label>
            <select
                name="application"
                value={query.application}
                onChange={handleChange}
                style={styles.select}
                required
            >
              <option value="">Select Application</option>
              {filters.map((app) => (
                  <option key={app.application} value={app.application}>
                    {app.application}
                  </option>
              ))}
            </select>
          </div>

          <div style={styles.formGroup}>
            <label style={styles.label}>Database</label>
            <select
                name="db_name"
                value={query.db_name}
                onChange={handleChange}
                style={styles.select}
                required
            >
              <option value="">Select Database</option>
              {query.application &&
                  getDatabases(query.application).map((db) => (
                      <option key={db.name} value={db.name}>
                        {db.name}
                      </option>
                  ))}
            </select>
          </div>

          <div style={styles.formGroup}>
            <label style={styles.label}>Schema</label>
            <select
                name="db_schema"
                value={query.db_schema}
                onChange={handleChange}
                style={styles.select}
                required
            >
              <option value="">Select Schema</option>
              {query.db_name &&
                  getSchemas(query.application, query.db_name).map((schema) => (
                      <option key={schema.name} value={schema.name}>
                        {schema.name}
                      </option>
                  ))}
            </select>
          </div>

          <div style={styles.formGroup}>
            <label style={styles.label}>Table</label>
            <select
                name="db_table"
                value={query.db_table}
                onChange={handleChange}
                style={styles.select}
                required
            >
              <option value="">Select Table</option>
              {query.db_schema &&
                  getTables(query.application, query.db_name, query.db_schema).map((table) => (
                      <option key={table} value={table}>
                        {table}
                      </option>
                  ))}
            </select>
          </div>

          <div style={styles.formGroup}>
            <label style={styles.label}>Operation</label>
            <select
                name="event_operation"
                value={query.event_operation}
                onChange={handleChange}
                style={styles.select}
                required
            >
              <option value="">Select Operation</option>
              <option value="C">Create</option>
              <option value="U">Update</option>
              <option value="D">Delete</option>
            </select>
          </div>

          <div style={styles.formGroup}>
            <label style={styles.label}>Start Date</label>
            <input
                type="datetime-local"
                name="start_date"
                value={query.start_date}
                onChange={handleChange}
                style={styles.input}
                required
            />
          </div>

          <div style={styles.formGroup}>
            <label style={styles.label}>End Date</label>
            <input
                type="datetime-local"
                name="end_date"
                value={query.end_date}
                onChange={handleChange}
                style={styles.input}
                required
            />
          </div>

          <button
              type="submit"
              style={{ ...styles.button, ...(loading && styles.buttonDisabled) }}
              disabled={loading}
          >
            {loading ? "Searching..." : "Search"}
          </button>
        </form>

        <div style={styles.tableContainer}>
          <h2>Results</h2>
          <table style={styles.table}>
            <thead>
            <tr>
              <th style={styles.th}>Application</th>
              <th style={styles.th}>Database</th>
              <th style={styles.th}>Schema</th>
              <th style={styles.th}>Table</th>
              <th style={styles.th}>Operation</th>
              <th style={styles.th}>Event Date</th>
              <th style={styles.th}>Details</th>
            </tr>
            </thead>
            <tbody>
            {results.length > 0 ? (
                results.map((result, index) => (
                    <React.Fragment key={index}>
                      <tr onClick={() => toggleRow(index)} style={{ cursor: "pointer" }}>
                        <td style={styles.td}>{result.application}</td>
                        <td style={styles.td}>{result.dbName}</td>
                        <td style={styles.td}>{result.dbSchema}</td>
                        <td style={styles.td}>{result.dbTable}</td>
                        <td style={styles.td}>{getOperationLabel(result.eventOperation)}</td>
                        <td style={styles.td}>{new Date(result.eventDate).toLocaleString()}</td>
                        <td style={styles.td}>View JSON</td>
                      </tr>
                      {expandedRow === index && (
                          <tr>
                            <td colSpan="7" style={styles.expandedRow}>
                              <SyntaxHighlighter language="json" style={coy}>
                                {JSON.stringify(JSON.parse(result.event), null, 2)}
                              </SyntaxHighlighter>
                            </td>
                          </tr>
                      )}
                    </React.Fragment>
                ))
            ) : (
                <tr>
                  <td colSpan="7" style={{ ...styles.td, ...styles.noResults }}>
                    No results found
                  </td>
                </tr>
            )}
            </tbody>
          </table>
        </div>
      </div>
  );
};

export default App;