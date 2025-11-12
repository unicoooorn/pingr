package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
	"github.com/unicoooorn/pingr/internal/service/mocks"
)

func TestKakDela(t *testing.T) {
	svc := service.New(nil, nil, nil, nil, nil, config.Config{})

	res, err := svc.GetStatus(context.Background(), "kak dela")

	assert.NoError(t, err)
	assert.Equal(t, model.CheckResult{Status: "ok", Details: ""}, res)
}

func TestInitiateCheck_AllHealthy_NoAlertSent(t *testing.T) {
	// Создаем моки
	checker := &mocks.MockChecker{}
	alertSender := &mocks.MockAlertSender{}
	alertGenerator := &mocks.MockAlertGenerator{}
	metricsExtractor := &mocks.MockMetricsExtractor{}
	infographicsRenderer := &mocks.MockInfographicsRenderer{}

	// Конфиг с двумя бэкендами
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"backend1": {},
			"backend2": {},
		},
	}

	// Создаем сервис
	srv := service.New(
		checker,
		alertSender,
		alertGenerator,
		metricsExtractor,
		infographicsRenderer,
		cfg,
	)

	// Настраиваем ожидания - все проверки возвращают здоровый статус
	checker.On("Check", mock.Anything, "backend1").
		Return(model.CheckResult{Status: model.PingStatusOk}, nil).Once()
	checker.On("Check", mock.Anything, "backend2").
		Return(model.CheckResult{Status: model.PingStatusOk}, nil).Once()

	// AlertSender.SendAlert НЕ должен вызываться
	alertSender.AssertNotCalled(t, "SendAlert")

	// Выполняем проверку
	err := srv.InitiateCheck(context.Background())

	// Проверяем что нет ошибок и все моки вызваны как ожидалось
	assert.NoError(t, err)
	checker.AssertExpectations(t)
	alertSender.AssertExpectations(t)
}

func TestInitiateCheck_UnhealthySystem_AlertSent(t *testing.T) {
	// Создаем моки
	checker := &mocks.MockChecker{}
	alertSender := &mocks.MockAlertSender{}
	alertGenerator := &mocks.MockAlertGenerator{}
	metricsExtractor := &mocks.MockMetricsExtractor{}
	infographicsRenderer := &mocks.MockInfographicsRenderer{}

	// Конфиг с двумя бэкендами
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"backend1": {},
			"backend2": {},
		},
	}

	// Создаем сервис
	srv := service.New(
		checker,
		alertSender,
		alertGenerator,
		metricsExtractor,
		infographicsRenderer,
		cfg,
	)

	// Настраиваем ожидания - один бэкенд нездоров
	checker.On("Check", mock.Anything, "backend1").
		Return(model.CheckResult{Status: model.PingStatusNotOk}, nil).Once()
	checker.On("Check", mock.Anything, "backend2").
		Return(model.CheckResult{Status: model.PingStatusOk}, nil).Once()

	// Метрики для обоих бэкендов
	metricsExtractor.On("Extract", mock.Anything, "backend1").
		Return(model.MetricsExtractorResult{}, nil).Once()
	metricsExtractor.On("Extract", mock.Anything, "backend2").
		Return(model.MetricsExtractorResult{}, nil).Once()

	// Генерация сообщения алерта
	alertGenerator.On("GenerateAlertMessage", mock.Anything, mock.AnythingOfType("map[string]model.SubsystemInfo")).
		Return("Test alert message", nil).Once()

	// Рендер инфографики
	infographicsRenderer.On("Render", mock.Anything, mock.AnythingOfType("map[string]model.SubsystemInfo")).
		Return([]byte("infographic"), nil).Once()

	// Отправка алерта
	alertSender.On("SendAlert", mock.Anything, "Test alert message", []byte("infographic")).
		Return(nil).Once()

	// Выполняем проверку
	err := srv.InitiateCheck(context.Background())

	// Проверяем что нет ошибок и все моки вызваны как ожидалось
	assert.NoError(t, err)
	checker.AssertExpectations(t)
	metricsExtractor.AssertExpectations(t)
	alertGenerator.AssertExpectations(t)
	infographicsRenderer.AssertExpectations(t)
	alertSender.AssertExpectations(t)
}

// Тест когда проверка возвращает ошибку
func TestInitiateCheck_CheckFails_ReturnsError(t *testing.T) {
	checker := &mocks.MockChecker{}
	alertSender := &mocks.MockAlertSender{}
	// Остальные моки не понадобятся т.к. мы не дойдем до алертинга

	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"backend1": {},
		},
	}

	srv := service.New(
		checker,
		alertSender,
		&mocks.MockAlertGenerator{},
		&mocks.MockMetricsExtractor{},
		&mocks.MockInfographicsRenderer{},
		cfg,
	)

	// Проверка завершается ошибкой
	expectedErr := errors.New("check failed")
	checker.On("Check", mock.Anything, "backend1").
		Return(model.CheckResult{}, expectedErr).Once()

	err := srv.InitiateCheck(context.Background())

	// Должны получить ошибку
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check stage")
	assert.Contains(t, err.Error(), "check failed")
	checker.AssertExpectations(t)
}

// Тест для GetStatus с специальным случаем
func TestGetStatus_SpecialCase(t *testing.T) {
	srv := service.New(
		&mocks.MockChecker{},
		&mocks.MockAlertSender{},
		&mocks.MockAlertGenerator{},
		&mocks.MockMetricsExtractor{},
		&mocks.MockInfographicsRenderer{},
		config.Config{},
	)

	result, err := srv.GetStatus(context.Background(), "kak dela")

	assert.NoError(t, err)
	assert.Equal(t, model.CheckResult{Status: "ok", Details: ""}, result)
}
