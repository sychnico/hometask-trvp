// src/App.js
import React from "react";
import { BrowserRouter as Router, Routes, Route, useLocation } from "react-router-dom";
import ReferenceForm from "./components/ReferenceForm";
import ArticleSearch from "./components/ArticleSearch";
import HomePage from "./components/HomePage";
import ReferenceFormMultiRow from "./components/ReferenceFormMultiRow";

function ReferenceFormAnswerWrapper() {
  const location = useLocation();
  // ✅ Извлекаем initialAnswer из state
  const initialAnswer = location.state?.initialAnswer || ""; 
  
  // ✅ Передаем initialAnswer в ReferenceForm
  return <ReferenceFormMultiRow initialAnswer={initialAnswer} />; 
}

function App() {
  return (
    <Router>
      <Routes>
        {/* Главная страница */}
        <Route path="/" element={<HomePage />} /> 
        
        {/* Страница стандартной формы */}
        <Route path="/reference-form" element={<ReferenceForm />} /> 
        
        {/* Страница многострочной формы */}
        <Route path="/reference-form-multi-row" element={<ReferenceFormAnswerWrapper />} />

        <Route path="/search" element={<ArticleSearch />} />
      </Routes>
      
    </Router>
  );
}

export default App;
