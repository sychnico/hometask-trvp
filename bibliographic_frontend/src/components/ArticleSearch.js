// src/components/ArticleSearch.js
import React, { useState } from "react";
import {
  Container,
  Typography,
  TextField,
  Button,
  CircularProgress,
  Box,
  List,
  ListItem,
  ListItemText,
  Link,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import { REF_FORM_MULTYROW_URL, SEARCH_ELIBRARY_URL } from '../consts';

function ArticleSearch() {
  const [searchTerm, setSearchTerm] = useState("");
  const [articles, setArticles] = useState([]); 
  const [selectedArticles, setSelectedArticles] = useState(new Set()); 
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSearch = async (e) => {
    e.preventDefault();
    setArticles([]);
    setSelectedArticles(new Set());
    setError("");
    setLoading(true);

    try {
      const res = await fetch(`${SEARCH_ELIBRARY_URL}?query=${encodeURIComponent(searchTerm)}`);

      if (!res.ok) {
        const text = await res.text();
        setError('Что-то пошло не так')
        console.log(`Ошибка: ${res.status} - ${text}`);
        return;
      }

      const data = await res.json();

      if (data) {
        setArticles(data);
      } else {
        setArticles([]);
        setError('Поиск не дал результатов');
      }
    } catch (err) {
      setError("Ошибка сети: сервер на обслуживании");
      console.log("Ошибка сети: " + err.message);
    } finally {
      setLoading(false);
    }
  };

  // ФУНКЦИЯ ДЛЯ ПЕРЕКЛЮЧЕНИЯ ФЛАЖКА
  const handleToggle = (link, title) => (e) => {
    const key = title + link; 
    const newSelected = new Set(selectedArticles); 
    
    if (e.target.checked) {
      newSelected.add(key);
    } else {
      newSelected.delete(key);
    }
    
    setSelectedArticles(newSelected);
  };

  // функция для генерации ссылок по выбраным источникам
  const handleGenerateReferences = async () => {
    if (selectedArticles.size === 0) {
      alert("Выберите хотя бы одну статью.");
      return;
    }

    setLoading(true);
    setError("");
    const links = [...selectedArticles]
              .map((item, index) => `${item}`)
              .join('\n ');

    const payload = {
      user_request: links, 
      prompt_type: "Статья из журнала", // ищем именно в elibrary, поэтому тип определен заранее
      example_record: null,
    };

    try {
      const res = await fetch(`${REF_FORM_MULTYROW_URL}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const text = await res.text();
        setError('Что-то пошло не так')
        console.log(`Ошибка при генерации: ${res.status} - ${text}`);
        setLoading(false);
        return;
      }

      const data = await res.json();
      const generatedAnswer = data.answer || "Библиографические ссылки сгенерированы.";
      
      navigate("/reference-form-multi-row", { state: { initialAnswer: generatedAnswer } });

    } catch (err) {
      console.log(err.message)
      setError("Ошибка сети: сервер на обслуживании");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="md" sx={{ mt: 5 }}>
      <Typography variant="h4" gutterBottom align="center">
        Поиск статей в e-library
      </Typography>

      <Box sx={{ mb: 2 }}>
         <Button 
            variant="text" 
            onClick={() => navigate("/reference-form-multi-row")}
         >
            ← Вернуться к форме
         </Button>
      </Box>

      <Box component="form" onSubmit={handleSearch} sx={{ display: "flex", gap: 2, mb: 3 }}>
        <TextField
          label="Введите ключевые слова для поиска"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          fullWidth
          required
        />
        <Button type="submit" variant="contained" disabled={loading}>
          Поиск
        </Button>
      </Box>

      {articles.length > 0 && (
        <>

          <Button
            variant="contained"
            size="large"
            fullWidth
            onClick={handleGenerateReferences}
            disabled={selectedArticles.size === 0 || loading} 
            sx={{ mt: 3, mb: 2 }}
          >
            Оформить выбранные источники ({selectedArticles.size})
          </Button>

          <Box sx={{ 
            width: "100%", 
            p: 2, 
            mb: 3, 
            backgroundColor: '#f5f5f5', 
            borderRadius: 1,
            borderLeft: '4px solid #1976d2',
          }}>
            <Typography variant="subtitle1" component="h2" sx={{ fontWeight: 'bold', mb: 1, color: '#1976d2' }}>
              Подсказка:
            </Typography>
            <Typography variant="body2" sx={{ mb: 1 }}>
              Выберите нужные Вам статьи из списка ниже, нажимая на квадратик сбоку от названия.
              После чего нажмите на кнопку <i>"Оформить выбранные источники"</i> и система оформит источники по ГОСТ за вас!
            </Typography>
          </Box>

          <List>
            {articles.map((article) => {
              const checkboxKey = article.title + article.link; 
              
              return (
                <ListItem 
                  key={article.link} 
                  sx={{ display: 'flex', alignItems: 'flex-start' }} 
                >
                  <input
                      type="checkbox"
                      checked={selectedArticles.has(checkboxKey)} 
                      onChange={handleToggle(article.link, article.title)} 
                      style={{
                        width: '24px',
                        height: '24px',
                        transform: 'scale(1.5)', // дополнительно масштабируем, если надо ещё больше
                        marginTop: '8px',
                        marginRight: '16px'
                      }}
                  />
                  
                  <ListItemText
                      primary={article.title}
                      secondary={
                          <Link href={article.link} target="_blank" rel="noopener" variant="body2">
                              {article.link}
                          </Link>
                      }
                  />
                </ListItem>
              )
            })}
          </List>
          
          

           <Button 
                variant="outlined" 
                fullWidth
                onClick={() => navigate("/reference-form-multi-row")}
                sx={{
                  borderWidth: 2,
                  borderStyle: 'solid',
                  backgroundColor: '#e0e0e0',
                  ':hover': {
                    backgroundColor: '#ddd5d5ff',
                  },
                }}
            >
                Вернуться к форме (отмена выбора)
           </Button>
        </>
      )}

      {loading && <CircularProgress sx={{ display: "block", mx: "auto" }} />}
      {error && <Typography color="error" sx={{ mt: 2 }}>{error}</Typography>}

    </Container>
  );
}

export default ArticleSearch;
