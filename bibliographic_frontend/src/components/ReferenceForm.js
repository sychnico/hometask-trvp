import React, { useState, useEffect } from "react";
import {
  Container,
  Typography,
  TextField,
  Button,
  MenuItem,
  CircularProgress,
  Box,
  Tooltip,
  IconButton,
  InputAdornment,
} from "@mui/material";
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import { useNavigate } from "react-router-dom"; 

function ReferenceForm({ initialAnswer = "" }) {
  const [userRequest, setUserRequest] = useState("");
  const [promptType, setPromptType] = useState("");
  const [customType, setCustomType] = useState("");
  const [exampleRecord, setExampleRecord] = useState("");
  const [answer, setAnswer] = useState(initialAnswer); 
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (initialAnswer) {
      setAnswer(initialAnswer);
    }
  }, [initialAnswer]); 


  const handleSubmit = async (e) => {
    e.preventDefault();
    setAnswer("");
    setError("");
    setLoading(true);

    const finalType = promptType === "Другой" ? customType || null : promptType || null;

    const payload = {
      user_request: userRequest,
      prompt_type: finalType,
      example_record: exampleRecord || null,
    };

   
    try {
        const res = await fetch("http://localhost:8080/request", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });

        if (!res.ok) {
            const text = await res.text();
            setError(`Ошибка: ${res.status} - ${text}`);
            setLoading(false);
            return;
        }

        const data = await res.json();
        setAnswer(data.answer); 
    } catch (err) {
        setError("Ошибка сети: " + err.message);
    } finally {
        setLoading(false);
    }
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 5, display: "flex", flexDirection: "column", alignItems: "center" }}>
      <Typography variant="h4" gutterBottom align="center">
        Составление библиографической ссылки
      </Typography>

      <Box component="form" onSubmit={handleSubmit} sx={{ width: "100%", display: "flex", flexDirection: "column", gap: 2 }}>
        
        <TextField
          label="Запрос пользователя"
          value={userRequest}
          onChange={(e) => setUserRequest(e.target.value)}
          required
          // ✅ Добавляем подсказку
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <Tooltip title="Введите информацию об источнике (например: 'статья иванова и и в журнале вестник науки')">
                  <IconButton edge="end">
                    <InfoOutlinedIcon />
                  </IconButton>
                </Tooltip>
              </InputAdornment>
            ),
          }}
        />
        
        {/* ✅ Заменяем FormControl/Select на TextField select для простоты добавления иконки */}
        <TextField
          select
          fullWidth
          label="Выбрать тип записи (или оставить по умолчанию)"
          value={promptType}
          onChange={(e) => setPromptType(e.target.value)}
          // ✅ Добавляем подсказку
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <Tooltip title="Выберите тип источника (книга, статья и т.д.) для более точного форматирования.">
                  {/* Оборачиваем IconButton в Box, чтобы он не перекрывал стрелку Select */}
                  <Box sx={{ mr: 2 }}> 
                    <IconButton edge="end">
                      <InfoOutlinedIcon />
                    </IconButton>
                  </Box>
                </Tooltip>
              </InputAdornment>
            ),
          }}
        >
          <MenuItem value=""><em>-- Выберите тип --</em></MenuItem>
          <MenuItem value="Книга">Книга</MenuItem>
          <MenuItem value="Интернет-ресурс">Интернет-ресурс</MenuItem>
          <MenuItem value="Закон, нормативный акт и т.п.">Закон, нормативный акт и т.п.</MenuItem>
          <MenuItem value="Диссертация">Диссертация</MenuItem>
          <MenuItem value="Автореферат">Автореферат</MenuItem>
          <MenuItem value="Статья из журнала">Статья из журнала</MenuItem>
          <MenuItem value="Статья из сборника">Статья из сборника</MenuItem>
          <MenuItem value="Статья из газеты">Статья из газеты</MenuItem>
          <MenuItem value="Другой">Другой</MenuItem>
        </TextField>

        {promptType === "Другой" && (
          <TextField
            label="Указать свой тип записи"
            value={customType}
            onChange={(e) => setCustomType(e.target.value)}
            // ✅ Добавляем подсказку
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title="Введите свой собственный тип источника, если его нет в списке.">
                    <IconButton edge="end">
                      <InfoOutlinedIcon />
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
          />
        )}

        <TextField
          label="Указать определенный формат записи (или оставить пустым по умолчанию)"
          value={exampleRecord}
          onChange={(e) => setExampleRecord(e.target.value)}
          // ✅ Добавляем подсказку
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <Tooltip title="Если вам нужен конкретный ГОСТ или стиль (например, 'ГОСТ Р 7.0.5-2008' или 'APA'), укажите его здесь.">
                  <IconButton edge="end">
                    <InfoOutlinedIcon />
                  </IconButton>
                </Tooltip>
              </InputAdornment>
            ),
          }}
        />

        <Button type="submit" variant="contained" size="large">
          Отправить
        </Button>
        
        <Button 
          variant="outlined" 
          size="large"
          onClick={() => navigate("/search")}
          sx={{
            borderWidth: 2,
            borderStyle: 'solid',
            mt: 1,
            backgroundColor: '#e0e0e0', // устанавливаем фоновое оформление
            ':hover': {               // применяем эффекты при наведении
              backgroundColor: '#ddd5d5ff', // цвет фона при наведении
            },
          }}
        >
          Найти статьи в e-library
        </Button>
      </Box>

      {/* Блок вывода ответа */}
      {loading && <CircularProgress sx={{ mt: 3 }} />}
      {error && <Typography color="error" sx={{ mt: 2 }}>{error}</Typography>}
      {answer && (
        <TextField
          value={answer}
          multiline
          rows={10}
          fullWidth
          InputProps={{ readOnly: true }}
          sx={{ mt: 3 }}
        />
      )}
    </Container>
  );
}

export default ReferenceForm;
