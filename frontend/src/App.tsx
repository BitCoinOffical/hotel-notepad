import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { ApiError, authApi, guestbookApi } from './api';
import type {
  CreateGuestRequest,
  Guest,
  GuestFormValues,
  RegisterRequest,
  UpdateGuestRequest,
} from './types';

type AuthMode = 'login' | 'register';
type Screen = 'checking' | 'auth' | 'app';

const EMPTY_FORM: GuestFormValues = {
  firstName: '',
  midName: '',
  lastName: '',
  note: '',
  phone: '',
  arrivalDate: '',
  departureDate: '',
};

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return 'Произошла неизвестная ошибка.';
}

function dateForInput(value: string): string {
  return value ? value.slice(0, 10) : '';
}

function formatDate(value: string): string {
  if (!value) return '—';
  const [year, month, day] = value.slice(0, 10).split('-');
  return `${day}.${month}.${year}`;
}

function toISODate(value: string): string {
  return `${value}T00:00:00Z`;
}

function guestToForm(guest: Guest): GuestFormValues {
  return {
    firstName: guest.first_name,
    midName: guest.mid_name,
    lastName: guest.last_name,
    note: guest.note ?? '',
    phone: String(guest.phone),
    arrivalDate: dateForInput(guest.arrival_date),
    departureDate: dateForInput(guest.departure_date),
  };
}

function App() {
  const [screen, setScreen] = useState<Screen>('checking');
  const [authMode, setAuthMode] = useState<AuthMode>('login');
  const [guests, setGuests] = useState<Guest[]>([]);
  const [search, setSearch] = useState('');
  const [pageError, setPageError] = useState('');
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editingGuest, setEditingGuest] = useState<Guest | null>(null);

  useEffect(() => {
    let active = true;

    guestbookApi.getEntries()
      .then((entries) => {
        if (!active) return;
        setGuests(entries);
        setScreen('app');
      })
      .catch((error: unknown) => {
        if (!active) return;
        if (error instanceof ApiError && error.status === 401) {
          setScreen('auth');
        } else {
          setScreen('auth');
          setPageError(getErrorMessage(error));
        }
      });

    return () => {
      active = false;
    };
  }, []);

  const filteredGuests = useMemo(() => {
    const value = search.trim().toLocaleLowerCase('ru');
    if (!value) return guests;

    return guests.filter((guest) => {
      const haystack = [
        guest.first_name,
        guest.mid_name,
        guest.last_name,
        guest.phone,
        guest.note ?? '',
      ].join(' ').toLocaleLowerCase('ru');
      return haystack.includes(value);
    });
  }, [guests, search]);

  async function refreshEntries(): Promise<void> {
    const entries = await guestbookApi.getEntries();
    setGuests(entries);
  }

  async function handleAuth(
    mode: AuthMode,
    username: string,
    password: string,
    passwordRetry: string,
  ): Promise<void> {
    setLoading(true);
    setPageError('');

    try {
      if (mode === 'register') {
        const data: RegisterRequest = {
          username,
          password,
          password_retry: passwordRetry,
        };
        await authApi.register(data);
      } else {
        await authApi.login({ username, password });
      }

      await refreshEntries();
      setScreen('app');
    } catch (error: unknown) {
      setPageError(getErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  async function handleLogout(): Promise<void> {
    setLoading(true);
    setPageError('');

    try {
      await authApi.logout();
      setGuests([]);
      setSearch('');
      setScreen('auth');
      setAuthMode('login');
    } catch (error: unknown) {
      setPageError(getErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  async function handleSaveGuest(values: GuestFormValues): Promise<void> {
    const common = {
      first_name: values.firstName.trim(),
      mid_name: values.midName.trim(),
      last_name: values.lastName.trim(),
      phone: Number(values.phone),
      arrival_date: toISODate(values.arrivalDate),
      departure_date: toISODate(values.departureDate),
    };

    setLoading(true);
    setPageError('');

    try {
      if (editingGuest) {
        const payload: UpdateGuestRequest = {
          ...common,
          // Пустая строка позволяет очистить заметку при редактировании.
          note: values.note,
        };
        await guestbookApi.updateEntry(editingGuest.id, payload);
      } else {
        const payload: CreateGuestRequest = {
          ...common,
          note: values.note.trim() ? values.note.trim() : null,
        };
        await guestbookApi.createEntry(payload);
      }

      await refreshEntries();
      setFormOpen(false);
      setEditingGuest(null);
    } catch (error: unknown) {
      setPageError(getErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(guest: Guest): Promise<void> {
    const fullName = `${guest.first_name} ${guest.last_name}`;
    if (!window.confirm(`Удалить запись «${fullName}»?`)) return;

    setLoading(true);
    setPageError('');

    try {
      await guestbookApi.deleteEntry(guest.id);
      await refreshEntries();
    } catch (error: unknown) {
      setPageError(getErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  if (screen === 'checking') {
    return (
      <main className="center-screen">
        <p className="muted">Проверяем сессию…</p>
      </main>
    );
  }

  if (screen === 'auth') {
    return (
      <AuthPage
        mode={authMode}
        loading={loading}
        error={pageError}
        onModeChange={(mode) => {
          setAuthMode(mode);
          setPageError('');
        }}
        onSubmit={handleAuth}
      />
    );
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">HOTEL NOTEBOOK</p>
          <h1>Гостевая книга</h1>
        </div>
        <button className="button button-secondary" type="button" onClick={handleLogout} disabled={loading}>
          Выйти
        </button>
      </header>

      <main className="main-content">
        <section className="list-heading">
          <div>
            <h2>Записи гостей</h2>
            <p className="muted">Всего записей: {guests.length}</p>
          </div>
          <button
            className="button button-primary"
            type="button"
            onClick={() => {
              setEditingGuest(null);
              setFormOpen(true);
              setPageError('');
            }}
            disabled={loading}
          >
            Добавить гостя
          </button>
        </section>

        {pageError && <div className="notice notice-error" role="alert">{pageError}</div>}

        {formOpen && (
          <GuestForm
            key={editingGuest?.id ?? 'new-guest'}
            initial={editingGuest ? guestToForm(editingGuest) : EMPTY_FORM}
            title={editingGuest ? 'Редактировать запись' : 'Новая запись'}
            loading={loading}
            onCancel={() => {
              setFormOpen(false);
              setEditingGuest(null);
            }}
            onSubmit={handleSaveGuest}
          />
        )}

        <section className="panel">
          <div className="search-row">
            <label className="search-label" htmlFor="guest-search">Поиск по имени, телефону или заметке</label>
            <input
              id="guest-search"
              className="input search-input"
              type="search"
              placeholder="Начните вводить запрос…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>

          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Гость</th>
                  <th>Телефон</th>
                  <th>Заезд</th>
                  <th>Выезд</th>
                  <th>Заметка</th>
                  <th className="actions-heading">Действия</th>
                </tr>
              </thead>
              <tbody>
                {filteredGuests.map((guest) => (
                  <tr key={guest.id}>
                    <td>
                      <strong>{guest.first_name} {guest.mid_name} {guest.last_name}</strong>
                    </td>
                    <td>{guest.phone}</td>
                    <td>{formatDate(guest.arrival_date)}</td>
                    <td>{formatDate(guest.departure_date)}</td>
                    <td className="note-cell">{guest.note || '—'}</td>
                    <td>
                      <div className="row-actions">
                        <button
                          className="text-button"
                          type="button"
                          disabled={loading}
                          onClick={() => {
                            setEditingGuest(guest);
                            setFormOpen(true);
                            setPageError('');
                          }}
                        >
                          Изменить
                        </button>
                        <button
                          className="text-button text-button-danger"
                          type="button"
                          disabled={loading}
                          onClick={() => void handleDelete(guest)}
                        >
                          Удалить
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {filteredGuests.length === 0 && (
              <div className="empty-state">
                <p className="empty-title">{search ? 'Ничего не найдено' : 'Записей пока нет'}</p>
                <p className="muted">
                  {search ? 'Попробуйте изменить поисковый запрос.' : 'Добавьте первую запись гостя.'}
                </p>
              </div>
            )}
          </div>
        </section>
      </main>
    </div>
  );
}

interface AuthPageProps {
  mode: AuthMode;
  loading: boolean;
  error: string;
  onModeChange: (mode: AuthMode) => void;
  onSubmit: (mode: AuthMode, username: string, password: string, passwordRetry: string) => Promise<void>;
}

function AuthPage({ mode, loading, error, onModeChange, onSubmit }: AuthPageProps) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [passwordRetry, setPasswordRetry] = useState('');
  const [validationError, setValidationError] = useState('');

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setValidationError('');

    if (mode === 'register' && password !== passwordRetry) {
      setValidationError('Пароли не совпадают.');
      return;
    }

    await onSubmit(mode, username.trim(), password, passwordRetry);
  }

  return (
    <main className="auth-screen">
      <section className="auth-card">
        <p className="eyebrow">HOTEL NOTEBOOK</p>
        <h1>{mode === 'login' ? 'Вход' : 'Регистрация'}</h1>
        <p className="muted auth-description">
          {mode === 'login' ? 'Войдите, чтобы открыть гостевую книгу.' : 'Создайте аккаунт для работы с гостевой книгой.'}
        </p>

        {(error || validationError) && (
          <div className="notice notice-error" role="alert">{validationError || error}</div>
        )}

        <form className="form-stack" onSubmit={handleSubmit}>
          <label>
            Имя пользователя
            <input
              className="input"
              type="text"
              autoComplete="username"
              minLength={2}
              maxLength={16}
              required
              value={username}
              onChange={(event) => setUsername(event.target.value)}
            />
          </label>
          <label>
            Пароль
            <input
              className="input"
              type="password"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              minLength={8}
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>
          {mode === 'register' && (
            <label>
              Повторите пароль
              <input
                className="input"
                type="password"
                autoComplete="new-password"
                minLength={8}
                required
                value={passwordRetry}
                onChange={(event) => setPasswordRetry(event.target.value)}
              />
            </label>
          )}
          <button className="button button-primary button-wide" type="submit" disabled={loading}>
            {loading ? 'Выполняется…' : mode === 'login' ? 'Войти' : 'Создать аккаунт'}
          </button>
        </form>

        <div className="auth-switch">
          {mode === 'login' ? 'Нет аккаунта?' : 'Уже есть аккаунт?'}{' '}
          <button
            className="text-button"
            type="button"
            onClick={() => onModeChange(mode === 'login' ? 'register' : 'login')}
          >
            {mode === 'login' ? 'Зарегистрироваться' : 'Войти'}
          </button>
        </div>
      </section>
    </main>
  );
}

interface GuestFormProps {
  initial: GuestFormValues;
  title: string;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: GuestFormValues) => Promise<void>;
}

function GuestForm({ initial, title, loading, onCancel, onSubmit }: GuestFormProps) {
  const [values, setValues] = useState<GuestFormValues>(initial);
  const [formError, setFormError] = useState('');

  function update<K extends keyof GuestFormValues>(key: K, value: GuestFormValues[K]) {
    setValues((current) => ({ ...current, [key]: value }));
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError('');

    if (!/^\d{11}$/.test(values.phone)) {
      setFormError('Телефон должен содержать ровно 11 цифр.');
      return;
    }

    if (values.departureDate < values.arrivalDate) {
      setFormError('Дата выезда не может быть раньше даты заезда.');
      return;
    }

    await onSubmit(values);
  }

  return (
    <section className="panel form-panel">
      <div className="panel-heading">
        <h3>{title}</h3>
        <button className="text-button" type="button" onClick={onCancel} disabled={loading}>
          Закрыть
        </button>
      </div>

      {formError && <div className="notice notice-error" role="alert">{formError}</div>}

      <form onSubmit={handleSubmit}>
        <div className="form-grid">
          <label>
            Имя
            <input className="input" value={values.firstName} onChange={(event) => update('firstName', event.target.value)} maxLength={255} required />
          </label>
          <label>
            Отчество
            <input className="input" value={values.midName} onChange={(event) => update('midName', event.target.value)} maxLength={255} required />
          </label>
          <label>
            Фамилия
            <input className="input" value={values.lastName} onChange={(event) => update('lastName', event.target.value)} maxLength={255} required />
          </label>
          <label>
            Телефон (11 цифр)
            <input className="input" type="tel" inputMode="numeric" value={values.phone} onChange={(event) => update('phone', event.target.value.replace(/\D/g, '').slice(0, 11))} maxLength={11} required />
          </label>
          <label>
            Дата заезда
            <input className="input" type="date" value={values.arrivalDate} onChange={(event) => update('arrivalDate', event.target.value)} required />
          </label>
          <label>
            Дата выезда
            <input className="input" type="date" value={values.departureDate} onChange={(event) => update('departureDate', event.target.value)} required />
          </label>
          <label className="full-width">
            Заметка <span className="muted inline-muted">(необязательно)</span>
            <textarea className="input textarea" value={values.note} onChange={(event) => update('note', event.target.value)} maxLength={2048} rows={3} />
          </label>
        </div>

        <div className="form-actions">
          <button className="button button-secondary" type="button" onClick={onCancel} disabled={loading}>Отмена</button>
          <button className="button button-primary" type="submit" disabled={loading}>
            {loading ? 'Сохранение…' : 'Сохранить'}
          </button>
        </div>
      </form>
    </section>
  );
}

export default App;
