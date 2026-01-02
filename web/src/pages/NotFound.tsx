import { Box, Typography, Button, Container } from '@mui/material';
import { Link } from 'react-router-dom';
import { Home } from '@mui/icons-material';

function NotFound() {
  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: '#1e1e1e',
      }}
    >
      <Container maxWidth="sm" sx={{ textAlign: 'center' }}>
        <Typography
          variant="h1"
          sx={{
            fontSize: '8rem',
            fontWeight: 'bold',
            color: '#90caf9',
            mb: 2,
          }}
        >
          404
        </Typography>
        <Typography
          variant="h4"
          sx={{ color: '#fff', mb: 1 }}
        >
          页面未找到
        </Typography>
        <Typography
          variant="body1"
          sx={{ color: 'text.secondary', mb: 4 }}
        >
          “休斯顿，我们有麻烦了！”
        </Typography>
        <Button
          variant="contained"
          startIcon={<Home />}
          component={Link}
          to="/"
          size="large"
        >
          返回首页
        </Button>
      </Container>
    </Box>
  );
}

export default NotFound;
