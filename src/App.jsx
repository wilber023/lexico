// JavaAnalyzer.jsx
import { useState } from 'react'; 
import './App.css'; // Asegúrate de tener un archivo CSS para estilos

const API_URL = 'http://localhost:8080/analyze';

const App = () => {
  const [code, setCode] = useState(`public class HolaMundo {
    public static void main(String[] args) {
        System.out.println("Hola Mundo desde Java!");

        int numero = 42;
        String nombre = "Java";

        for (int i = 0; i < 3; i++) {
            System.out.println("Iteración: " + i);
        }
    }
}`);
  const [analysis, setAnalysis] = useState(null);
  const [loading, setLoading] = useState(false);
  const [activeStep, setActiveStep] = useState('');

  const handleCodeChange = (e) => {
    setCode(e.target.value);
    setAnalysis(null);
    setActiveStep('');
  };

  const escapeHtml = (text) => {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  };

  const performLexicalAnalysis = async () => {
    if (!code.trim()) return alert('Por favor ingresa código Java para analizar');
    setLoading(true);
    setAnalysis(null);
    setActiveStep('');
    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code }),
      });

      if (!response.ok) throw new Error(`Error HTTP: ${response.status}`);
      const result = await response.json();
      setAnalysis(result);
      setActiveStep('lex');

    } catch (error) {
      console.error(error);
      alert('Error al conectar con el servidor. Asegúrate de que la API esté corriendo en :8080');
    } finally {
      setLoading(false);
    }
  };

  const renderStats = (stats) => (
    <div className="stats-grid">
      {['total_tokens', 'keywords', 'identifiers', 'symbols', 'numbers', 'strings'].map((key, idx) => (
        <div className="stat-card" key={idx}>
          <h3>{stats[key]}</h3>
          <p>{{
            total_tokens: 'Total Tokens',
            keywords: 'Palabras Reservadas',
            identifiers: 'Identificadores',
            symbols: 'Símbolos',
            numbers: 'Números',
            strings: 'Cadenas'
          }[key]}</p>
        </div>
      ))}
    </div>
  );

  const renderTable = (items, columns = '') => (
    <table>
      <thead>
        <tr>
          {columns.map((col, idx) => <th key={idx}>{col}</th>)}
        </tr>
      </thead>
      <tbody>
        {items.map((item, idx) => (
          <tr key={idx}>
            {columns.map((col, colIdx) => (
              <td key={colIdx}>
                {col === '#' ? idx + 1 :
                 col === 'Tipo' ? <span className={`token-type token-${item.type}`}>{item.type}</span> :
                 col === 'Valor' ? <code>{escapeHtml(item.value)}</code> :
item[col] ?? item[col.toLowerCase()] ?? ''
                 }
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );

  return (
    <div className="container">
      <header className="header">
        <h1>🔍 Analizador de Código Java</h1>
      </header>

      <main className="main-content">
        <div className="input-section">
          <label htmlFor="codeInput">Ingresa tu código Java:</label>
          <textarea
            id="codeInput"
            className="code-input"
            value={code}
            onChange={handleCodeChange}
            placeholder='public class MiClase { ... }'
          />
        </div>

        <div className="button-section">
          <button className="btn btn-primary" onClick={performLexicalAnalysis}>Análisis Léxico</button>
          <button className="btn btn-primary" onClick={() => setActiveStep('syn')} disabled={!analysis || (analysis.lex_errors?.length)}>Análisis Sintáctico</button>
          <button className="btn btn-primary" onClick={() => setActiveStep('sem')} disabled={!analysis || (analysis.syn_errors?.length)}>Análisis Semántico</button>
        </div>

        {loading && (
          <div className="loading">
            <div className="spinner"></div>
            Analizando código...
          </div>
        )}

        {analysis && (
          <>
            <section className="stats-section">{renderStats(analysis.stats)}</section>

            <section className="results-layout">
  {/* Columna izquierda: Panel de errores */}
  <div className="errors-panel">
    {activeStep === 'lex' && analysis.lex_errors?.length > 0 && (
      <div className="error-section">
        <div className="error-header">❌ Errores Léxicos</div>
        <div className="table-container">
           {renderTable(
  analysis.lex_errors.map(e => ({ 'Línea': e.line, 'Error': e.error })),
  ['Línea', 'Error']
)}


        </div>
      </div>
    )}

    {activeStep === 'syn' && analysis.syn_errors?.length > 0 && (
      <div className="error-section">
        <div className="error-header">⚠️ Errores Sintácticos</div>
        <div className="table-container">
{renderTable(
  analysis.syn_errors.map(e => ({ 'Línea': e.line, 'Error': e.error })),
  ['Línea', 'Error']
)}        </div>
      </div>
    )}

    {activeStep === 'sem' && analysis.sem_errors?.length > 0 && (
      <div className="error-section">
        <div className="error-header">🔍 Errores Semánticos</div>
        <div className="table-container">
{renderTable(
  analysis.sem_errors.map(e => ({ 'Línea': e.line, 'Error': e.error })),
  ['Línea', 'Error']
)}        </div>
      </div>
    )}

    {activeStep === 'sem' && (!analysis.sem_errors || analysis.sem_errors.length === 0) && (
      <div className="success-message">
        ✅ ¡Código analizado correctamente! No se encontraron errores.
      </div>
    )}
  </div>
 
  <div className="tokens-panel">
    <div className="result-header">📝 Tokens Identificados</div>
    <div className="table-container">
      {renderTable(analysis.tokens, ['#', 'Tipo', 'Valor', 'Línea'])}
    </div>
  </div>
</section>

          </>
        )}
      </main>
    </div>
  );
};

export default App;
