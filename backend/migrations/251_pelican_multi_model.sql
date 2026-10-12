ALTER TABLE pelican_test_plans ADD COLUMN model_ids text[];
UPDATE pelican_test_plans SET model_ids=ARRAY[model_id];
ALTER TABLE pelican_test_plans ALTER COLUMN model_ids SET NOT NULL;

-- Preserve compatibility with older single-model writers while maintaining the
-- first-model alias and validating the ordered list for every database writer.
CREATE FUNCTION pelican_plan_model_ids_sync() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE model text; seen text[] := ARRAY[]::text[];
BEGIN
  IF NEW.model_ids IS NULL THEN
    NEW.model_ids := ARRAY[NEW.model_id];
  ELSIF TG_OP='UPDATE' AND NEW.model_ids IS NOT DISTINCT FROM OLD.model_ids
    AND NEW.model_id IS DISTINCT FROM OLD.model_id THEN
    NEW.model_ids := ARRAY[NEW.model_id];
  END IF;
  IF array_ndims(NEW.model_ids) IS DISTINCT FROM 1
    OR cardinality(NEW.model_ids) NOT BETWEEN 1 AND 10 THEN
    RAISE EXCEPTION 'invalid pelican model_ids';
  END IF;
  FOREACH model IN ARRAY NEW.model_ids LOOP
    IF model IS NULL OR btrim(model)='' OR model<>btrim(model)
      OR length(model)>200 OR model=ANY(seen)
      OR position(chr(10) in model)>0 OR position(chr(13) in model)>0 THEN
      RAISE EXCEPTION 'invalid pelican model_ids';
    END IF;
    seen := array_append(seen, model);
  END LOOP;
  NEW.model_id := NEW.model_ids[array_lower(NEW.model_ids,1)];
  RETURN NEW;
END;
$$;
CREATE TRIGGER pelican_plan_model_ids_sync BEFORE INSERT OR UPDATE
ON pelican_test_plans FOR EACH ROW EXECUTE FUNCTION pelican_plan_model_ids_sync();
CREATE INDEX idx_pelican_results_plan_group_model_time
ON pelican_test_results(plan_id,group_id,model_id,finished_at DESC,id DESC);
