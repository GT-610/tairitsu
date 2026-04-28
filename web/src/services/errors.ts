import axios, { type AxiosError } from 'axios';
import { translateMessageCode } from '../i18n';

interface ErrorResponseData {
  error?: string;
  message?: string;
  error_code?: string;
  message_code?: string;
}

export function toError(error: unknown): Error {
  if (error instanceof Error) {
    return error;
  }

  return new Error(String(error));
}

export function isAxiosError<T = ErrorResponseData>(error: unknown): error is AxiosError<T> {
  return axios.isAxiosError<T>(error);
}

export function getErrorMessage(error: unknown, fallback: string): string {
  if (isAxiosError(error)) {
    const responseCode = error.response?.data?.error_code ?? error.response?.data?.message_code;
    if (typeof responseCode === 'string' && responseCode.trim() !== '') {
      return translateMessageCode(responseCode) ?? responseCode;
    }

    const responseMessage = error.response?.data?.error ?? error.response?.data?.message;
    if (typeof responseMessage === 'string' && responseMessage.trim() !== '') {
      return responseMessage;
    }

    if (typeof error.message === 'string' && error.message.trim() !== '') {
      return error.message;
    }
  }

  if (error instanceof Error && error.message.trim() !== '') {
    return error.message;
  }

  return fallback;
}

export function hasStatus(error: unknown, status: number): boolean {
  return isAxiosError(error) && error.response?.status === status;
}
