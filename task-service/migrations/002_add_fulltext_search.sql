-- колонка для поискового вектора
ALTER TABLE "Tasks" ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Создание индекса GIN для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_tasks_search_vector ON "Tasks" USING gin(search_vector);

-- Функция для обновления поискового вектора
CREATE OR REPLACE FUNCTION tasks_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector = 
        setweight(to_tsvector('russian', COALESCE(NEW.text, '')), 'A');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Триггер для автоматического обновления при вставке/обновлении
DROP TRIGGER IF EXISTS tasks_search_vector_trigger ON "Tasks";
CREATE TRIGGER tasks_search_vector_trigger 
    BEFORE INSERT OR UPDATE ON "Tasks" 
    FOR EACH ROW EXECUTE FUNCTION tasks_search_vector_update();

-- Обновление существующих записей
UPDATE "Tasks" SET search_vector = 
    setweight(to_tsvector('russian', COALESCE(text, '')), 'A');