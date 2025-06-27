import React, { useState } from 'react';

export function App() {
  const [code, setCode] = useState(`def factorial(n):
    if n <= 1:
        return 1
    else:
        return n * factorial(n - 1)

x = 5
print("El factorial de", x, "es", factorial(x))`);
  
  const [analysis, setAnalysis] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const analyzeCode = async () => {
    setLoading(true);
    setError('');
    
    try {
      const response = await fetch('http://localhost:8080/analyze', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code }),
      });
      
      if (!response.ok) {
        throw new Error('Error en el análisis');
      }
      
      const result = await response.json();
      setAnalysis(result);
    } catch (err) {
      setError('Error al conectar con el servidor: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold text-center mb-8 text-gray-800">
          Analizador Léxico, Sintáctico y Semántico de Factorial
        </h1>
        
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4 text-gray-700">Código a Analizar</h2>
            <textarea
              value={code}
              onChange={(e) => setCode(e.target.value)}
              className="w-full h-64 p-3 border border-gray-300 rounded-md font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Ingresa tu código aquí..."
            />
            <button
              onClick={analyzeCode}
              disabled={loading}
              className="mt-4 w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400 transition-colors"
            >
              {loading ? 'Analizando...' : 'Analizar Código'}
            </button>
            <button
            onClick={analyzeCode}
            disabled={loading}
            className="mt-4 w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400 transition-colors">
                sintactico
            </button>
            <button
            onClick={analyzeCode}
            disabled={loading}
            
            className="mt-4 w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400 transition-colors">
               semántico
            </button>
            {error && (
              <div className="mt-3 p-3 bg-red-100 border border-red-400 text-red-700 rounded-md">
                {error}
              </div>
            )}
          </div>


          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4 text-gray-700">Resultados del Análisis</h2>
            {analysis ? (
              <div className="space-y-6">
              
                <div>
                  <h3 className="text-lg font-medium mb-3 text-green-700">Análisis Léxico</h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm border-collapse border border-gray-300">
                      <thead>
                        <tr className="bg-gray-100">
                          <th className="border border-gray-300 px-3 py-2 text-left">Tipo</th>
                          <th className="border border-gray-300 px-3 py-2 text-left">Cantidad</th>
                          <th className="border border-gray-300 px-3 py-2 text-left">Tokens</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr>
                          <td className="border border-gray-300 px-3 py-2 font-medium">Palabras Reservadas</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.reserved_words.count}</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.reserved_words.tokens.join(', ')}</td>
                        </tr>
                        <tr>
                          <td className="border border-gray-300 px-3 py-2 font-medium">Identificadores</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.identifiers.count}</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.identifiers.tokens.join(', ')}</td>
                        </tr>
                        <tr>
                          <td className="border border-gray-300 px-3 py-2 font-medium">Números</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.numbers.count}</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.numbers.tokens.join(', ')}</td>
                        </tr>
                        <tr>
                          <td className="border border-gray-300 px-3 py-2 font-medium">Símbolos</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.symbols.count}</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.symbols.tokens.join(', ')}</td>
                        </tr>
                        <tr>
                          <td className="border border-gray-300 px-3 py-2 font-medium">Cadenas</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.strings.count}</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.strings.tokens.join(', ')}</td>
                        </tr>
                        <tr>
                          <td className="border border-gray-300 px-3 py-2 font-medium">Errores</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.errors.count}</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.errors.tokens.join(', ')}</td>
                        </tr>
                        <tr className="bg-gray-50 font-bold">
                          <td className="border border-gray-300 px-3 py-2">Total de Tokens</td>
                          <td className="border border-gray-300 px-3 py-2">{analysis.lexical.total_tokens}</td>
                          <td className="border border-gray-300 px-3 py-2">-</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>

            
                <div>
                  <h3 className="text-lg font-medium mb-3 text-blue-700">Análisis Sintáctico</h3>
                  <div className={`p-3 rounded-md ${analysis.syntax.valid ? 'bg-green-100 border border-green-400' : 'bg-red-100 border border-red-400'}`}>
                    <p className={`font-medium ${analysis.syntax.valid ? 'text-green-700' : 'text-red-700'}`}>
                      Estado: {analysis.syntax.valid ? 'Válido' : 'Inválido'}
                    </p>
                    <p className="mt-2 text-sm text-gray-600">{analysis.syntax.message}</p>
                    {analysis.syntax.errors && analysis.syntax.errors.length > 0 && (
                      <div className="mt-2">
                        <p className="text-sm font-medium text-red-600">Errores encontrados:</p>
                        <ul className="list-disc list-inside text-sm text-red-600">
                          {analysis.syntax.errors.map((error, index) => (
                            <li key={index}>{error}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                </div>

               
                <div>
                  <h3 className="text-lg font-medium mb-3 text-purple-700">Análisis Semántico</h3>
                  <div className={`p-3 rounded-md ${analysis.semantic.valid ? 'bg-green-100 border border-green-400' : 'bg-red-100 border border-red-400'}`}>
                    <p className={`font-medium ${analysis.semantic.valid ? 'text-green-700' : 'text-red-700'}`}>
                      Estado: {analysis.semantic.valid ? 'Válido' : 'Inválido'}
                    </p>
                    <p className="mt-2 text-sm text-gray-600">{analysis.semantic.message}</p>
                    {analysis.semantic.errors && analysis.semantic.errors.length > 0 && (
                      <div className="mt-2">
                        <p className="text-sm font-medium text-red-600">Errores encontrados:</p>
                        <ul className="list-disc list-inside text-sm text-red-600">
                          {analysis.semantic.errors.map((error, index) => (
                            <li key={index}>{error}</li>
                          ))}
                        </ul>
                      </div>
                    )}
               
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-500 py-8">
                <p>Ingresa código y haz clic en "Analizar Código" para ver los resultados</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}