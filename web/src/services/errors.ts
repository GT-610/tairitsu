import axios, { type AxiosError } from 'axios';
import { getDetailSeparator, translateMessageCode } from '../i18n';

interface ErrorResponseData {
  message?: string;
  error_code?: string;
  detail?: string;
}

export function toError(error: unknown): Error {
  if (error instanceof Error) {
    return error;
  }

  return new Error(String(error));
}

function isAxiosError<T = ErrorResponseData>(error: unknown): error is AxiosError<T> {
  return axios.isAxiosError<T>(error);
}

export function getErrorMessage(error: unknown, fallback: string): string {
  const sep = getDetailSeparator();

  if (isAxiosError(error)) {
    const responseCode = error.response?.data?.error_code;
    const responseDetail = error.response?.data?.detail;

    if (typeof responseCode === 'string' && responseCode.trim() !== '') {
      const translatedMessage = translateMessageCode(responseCode);
      if (translatedMessage) {
        if (typeof responseDetail === 'string' && responseDetail.trim() !== '') {
          return `${translatedMessage}${sep}${responseDetail}`;
        }
        return translatedMessage;
      }
    }

    const responseMessage = error.response?.data?.message;
    if (typeof responseMessage === 'string' && responseMessage.trim() !== '') {
      if (typeof responseDetail === 'string' && responseDetail.trim() !== '') {
        return `${responseMessage}${sep}${responseDetail}`;
      }
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
