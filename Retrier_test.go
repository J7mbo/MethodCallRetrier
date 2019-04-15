package MethodCallRetrier

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RetrierTestSuite struct {
	suite.Suite

	retrier *MethodCallRetrier
}

func (s *RetrierTestSuite) SetupTest() {
	s.retrier = New(0, 1, 1)
}

func TestRetrierTestSuite(t *testing.T) {
	suite.Run(t, new(RetrierTestSuite))
}

func (s *RetrierTestSuite) TestRetrierWorksWithPointer() {
	arg := "TestArg"

	results, _, _ := s.retrier.ExecuteWithRetry(&RetryObject{}, "MethodReturningString", arg)

	s.EqualValues(results[0], arg)
}

func (s *RetrierTestSuite) TestRetrierWorksWithObject() {
	arg := "TestArg"

	results, _, _ := s.retrier.ExecuteWithRetry(RetryObject{}, "MethodReturningString", arg)

	s.EqualValues(results[0], arg)
}

func (s *RetrierTestSuite) TestRetrierReturnsErrorOnInvalidMethod() {
	results, errs, _ := s.retrier.ExecuteWithRetry(RetryObject{}, "InvalidMethodName")

	s.Nil(results)
	s.Error(errs[0])
}

func (s *RetrierTestSuite) TestRetrierThrowsErrorReturnsNilResults() {
	results, _, _ := s.retrier.ExecuteWithRetry(RetryObject{}, "MethodReturningError", "TestArg")

	s.Nil(results)
}

func (s *RetrierTestSuite) TestRetrierThrowsErrorReturnsErrors() {
	_, errs, _ := s.retrier.ExecuteWithRetry(RetryObject{}, "MethodReturningError", "TestArg")

	s.IsType(errors.New(""), errs[0])
}

func (s *RetrierTestSuite) TestRetrierThrowsErrorReturnsCorrectNumberOfErrors() {
	_, errs, _ := s.retrier.ExecuteWithRetry(RetryObject{}, "MethodReturningError", "TestArg")

	s.Len(errs, 2)
}

func (s *RetrierTestSuite) TestRetrierReturnsNilWhenGivenObjectWithNoReturnTypes() {
	results, _, _ := s.retrier.ExecuteWithRetry(RetryObject{}, "MethodReturningNoValues")

	s.Len(results, 0)
}

func (s *RetrierTestSuite) TestRetrierRetriesCorrectNumberOfTimes() {
	testObj := RetryMockObject{}
	methodName := "MethodReturningError"

	testObj.On(methodName, "").Return(errors.New(""))

	_, _, _ = New(0, 5, 1).ExecuteWithRetry(&testObj, methodName, "")

	testObj.AssertNumberOfCalls(s.T(), methodName, 5)

	testObj.AssertExpectations(s.T())
}

func (s *RetrierTestSuite) TestRetrierWorksWithNegativeMaxRetries() {
	arg := "testArg"

	results, _, _ := New(-1, -1, 1).ExecuteWithRetry(RetryObject{}, "MethodReturningString", arg)

	s.EqualValues(results[0], arg)
}

func (s *RetrierTestSuite) TestRetrierDefaultsToOneRetryGivenZeroMaxRetries() {
	testObj := RetryMockObject{}

	New(0, 0, 1).ExecuteFuncWithRetry(testObj.MethodToBeCalledToReturnErrorWithTimesCalledAvailable)

	s.Equal(1, testObj.timesCalled)
}

func (s *RetrierTestSuite) TestRetrierReturnsAllErrorsPlusOurError() {
	testObj := RetryMockObject{}
	methodName := "MethodReturningError"

	testObj.On(methodName, "").Return(errors.New(""))

	_, errs, _ := New(0, 5, 1).ExecuteWithRetry(&testObj, methodName, "")

	s.Len(errs, 6)
}

func (s *RetrierTestSuite) TestRetrierWorksWhenErrorIsNotLastReturnParamOnObject() {
	testObj := RetryObject{}
	methodName := "MethodReturningErrorInRandomPosition"

	_, errs, _ := New(1, 1, 1).ExecuteWithRetry(&testObj, methodName, "")

	s.IsType(errors.New(""), errs[0])
}

func (s *RetrierTestSuite) TestRetrierWorksWhenMultipleReturnParamsAreErrors() {
	testObj := RetryObject{}
	methodName := "MethodReturningMultipleErrors"

	_, errs, _ := New(0, 5, 1).ExecuteWithRetry(&testObj, methodName, "")

	s.Len(errs, 11)
}

func (s *RetrierTestSuite) TestRetrierWorksWithUserFunction() {
	var num int

	errs, _ := New(0, 3, 1).ExecuteFuncWithRetry(func() error {
		num = 42

		return nil
	})

	s.Equal(42, num)
	s.Len(errs, 0)
}

func (s *RetrierTestSuite) TestRetrierWithUserFunctionReturnsCorrectNumberOfErrors() {
	errs, _ := New(0, 3, 1).ExecuteFuncWithRetry(func() error {
		return errors.New("")
	})

	s.Equal(4, len(errs))
}

func (s *RetrierTestSuite) TestRetrierWorksWithUserFunctionCalledCorrectNumberOfTimes() {
	testObj := RetryMockObject{}

	New(0, 3, 1).ExecuteFuncWithRetry(testObj.MethodToBeCalledToReturnErrorWithTimesCalledAvailable)

	s.Equal(3, testObj.timesCalled)
}

func (s *RetrierTestSuite) TestRetrierWithUserFunctionReturnsFalseWhenAllFailed() {
	testObj := RetryMockObject{}

	errs, wasSuccessful := New(0, 5, 1).ExecuteFuncWithRetry(testObj.MethodToBeCalledToReturnErrorWithTimesCalledAvailable)

	s.Equal(5, testObj.timesCalled)
	s.Equal(6, len(errs))
	s.False(wasSuccessful)
}

func (s *RetrierTestSuite) TestRetrierWithUserFunctionReturnsTrueWhenOneSucceededButOthersFailedFirst() {
	testObj := RetryMockObject{}
	resultStr := ""

	errs, wasSuccessful := New(0, 5, 1).ExecuteFuncWithRetry(func() error {
		result, err := testObj.MethodToBeCalledToReturnSuccessOnFifthCall()

		if err != nil {
			return err
		}

		resultStr = result

		return nil
	})

	s.Equal("omg", resultStr)
	s.Equal(5, testObj.timesCalled)
	s.Equal(4, len(errs))
	s.True(wasSuccessful)
}

func (s *RetrierTestSuite) TestRetrierWithMethodNameReturnsFalseWhenAllFailed() {
	testObj := RetryMockObject{}

	results, errs, wasSuccessful := New(0, 5, 1).ExecuteWithRetry(&testObj, "MethodToBeCalledToReturnErrorWithTimesCalledAvailable")

	s.Nil(results)
	s.Equal(5, testObj.timesCalled)
	s.Equal(6, len(errs))
	s.False(wasSuccessful)
}

func (s *RetrierTestSuite) TestRetrierWithMethodNameReturnsTrueWhenOneSucceededButOthersFailedFirst() {
	testObj := RetryMockObject{}

	results, errs, wasSuccessful := New(0, 5, 1).ExecuteWithRetry(&testObj, "MethodToBeCalledToReturnSuccessOnFifthCall")

	s.Equal("omg", results[0])
	s.Equal(5, testObj.timesCalled)
	s.Equal(4, len(errs))
	s.True(wasSuccessful)
}

/* This really only exists for coverage */
func (s *RetrierTestSuite) TestMaxRetriesError() {
	methodName := "AMethod"
	waitTime := "42"
	maxRetries := "52"

	err := MaxRetriesError{methodName: methodName, waitTime: 42, maxRetries: 52}

	s.Contains(err.Error(), methodName)
	s.Contains(err.Error(), waitTime)
	s.Contains(err.Error(), maxRetries)
}

type RetryObject struct{}

func (m *RetryObject) MethodReturningNoValues() {}

func (m *RetryObject) MethodReturningString(anArgument string) string {
	return anArgument
}

func (m *RetryObject) MethodReturningError(anArgument string) error {
	return errors.New("")
}

func (m *RetryObject) MethodReturningErrorInRandomPosition() (string, error, string) {
	return "", errors.New(""), ""
}

func (m *RetryObject) MethodReturningMultipleErrors() (string, error, error) {
	return "", errors.New(""), errors.New("")
}

type RetryMockObject struct {
	mock.Mock

	timesCalled int
}

func (m *RetryMockObject) MethodReturningError(anArgument string) error {
	return m.Called(anArgument).Error(0)
}

func (m *RetryMockObject) MethodToBeCalledToReturnErrorWithTimesCalledAvailable() error {
	m.timesCalled += 1

	return errors.New("")
}

func (m *RetryMockObject) MethodToBeCalledToReturnSuccessOnFifthCall() (string, error) {
	m.timesCalled += 1

	if m.timesCalled < 5 {
		return "", errors.New("ah crap")
	}

	return "omg", nil
}